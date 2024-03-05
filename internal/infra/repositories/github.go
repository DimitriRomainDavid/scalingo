package repositories

import (
	"errors"
	"net/http"
	"scalingo/internal/core/dto"
	conf "scalingo/internal/infra/config"
	"scalingo/internal/shared"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

const (
	EventsEndpoint      = "events"
	RepoListEndpoint    = "repositories"
	GithubVersionHeader = "X-GitHub-Api-Version"

	Authorization = "Authorization"
	Bearer        = "Bearer "

	CreateEvent       = "CreateEvent"
	RepositoryRefType = "repository"

	Since = "?since="
)

type Github struct {
	URL                    string
	Token                  string
	Version                string
	LatestCreatedRepoRetry int
	UseCredentials         bool
}

func ProvideGithub(config *conf.Config) *Github {
	return &Github{
		URL:                    config.GitHubURL,
		Token:                  config.GitHubToken,
		UseCredentials:         config.GitHubCredentials,
		Version:                config.GitHubVersion,
		LatestCreatedRepoRetry: config.LatestCreatedRepoRetry,
	}
}

type Repo struct {
	ID int `json:"id"`
}
type Payload struct {
	RefType string `json:"ref_type"`
}

type Event struct {
	Type    string   `json:"type"`
	Repo    *Repo    `json:"repo"`
	Payload *Payload `json:"payload"`
}

func (g *Github) GetLatestRepoID() (int, error) {
	for i := 0; i < g.LatestCreatedRepoRetry; i++ {
		log.Warnf("try %d on %d to fetch latest created repo ID", i, g.LatestCreatedRepoRetry)
		statusCode, b, err := httpRequest(&fasthttp.Client{}, g.Version, g.URL+EventsEndpoint, g.Token, g.UseCredentials)
		if err != nil {
			return 0, errors.New("error while getting latest repo id: " + strconv.Itoa(statusCode) + shared.Separator + err.Error())
		}

		events := make([]*Event, 0)
		err = jsoniter.Unmarshal(b, &events)
		if err != nil {
			return 0, errors.New("error while deserializing latest repo id: " + err.Error())
		}

		for _, event := range events {
			if event.Type == CreateEvent && event.Payload.RefType == RepositoryRefType {
				log.Warnf("CreateEvent for %s detected...", event.Payload.RefType)
				return event.Repo.ID, nil
			}
		}
	}

	return 0, errors.New("error while getting latest repo id: couldn't find latest ID")
}

type Owner struct {
	Login string `json:"login"`
}

type Repository struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	FullName     string `json:"full_name"`
	Owner        *Owner `json:"owner"`
	LanguagesURL string `json:"languages_url"`
	HTMLURL      string `json:"html_url"`
	URL          string `json:"url"`
	Description  any    `json:"description"`
	Fork         bool   `json:"fork"`
}

func (g *Github) GetRepositories(id int) ([]*dto.LatestCreatedRepo, error) {
	url := g.URL + RepoListEndpoint + Since + strconv.Itoa(id)
	statusCode, repoList, err := httpRequest(&fasthttp.Client{}, g.Version, url, g.Token, g.UseCredentials)
	if err != nil {
		return nil, errors.New("error while getting repositories: " + strconv.Itoa(statusCode) + shared.Separator + err.Error())
	}

	repositories := make([]*Repository, 0)
	err = jsoniter.Unmarshal(repoList, &repositories)
	if err != nil {
		return nil, errors.New("error while deserializing repositories: " + strconv.Itoa(statusCode) + shared.Separator + err.Error())
	}

	latestCreatedRepos := make([]*dto.LatestCreatedRepo, 0)

	for _, repository := range repositories {
		var assertedDesc string
		if repository.Description != nil {
			assertedDesc = repository.Description.(string)
		}

		if !repository.Fork { // excluding forks
			latestCreatedRepos = append(latestCreatedRepos, &dto.LatestCreatedRepo{
				ID:           repository.ID,
				Name:         repository.Name,
				FullName:     repository.FullName,
				Owner:        &dto.Owner{Login: repository.Owner.Login},
				URL:          repository.URL,
				LanguagesURL: repository.LanguagesURL,
				HTMLURL:      repository.HTMLURL,
				Description:  assertedDesc,
			})
		}
	}
	return latestCreatedRepos, nil
}

func (g *Github) GetRepositoryLanguages(fullURL string) (map[string]int, error) {
	statusCode, repoList, err := httpRequest(&fasthttp.Client{}, g.Version, fullURL, g.Token, g.UseCredentials)
	if err != nil {
		return map[string]int{},
			errors.New("error while getting repository languages: " + strconv.Itoa(statusCode) + shared.Separator + err.Error())
	}
	if statusCode == http.StatusNotFound || statusCode == http.StatusMovedPermanently {
		log.Warnf("Language not found for %s, skipping...", fullURL)
		return map[string]int{}, nil
	}

	var languages map[string]int
	err = jsoniter.Unmarshal(repoList, &languages)
	if err != nil {
		return map[string]int{},
			errors.New("error while deserializing repository languages: " + strconv.Itoa(statusCode) + shared.Separator + err.Error())
	}
	return languages, nil
}

type spdx struct {
	License map[string]string `json:"license"`
}

func (g *Github) GetRepositorySPDX(fullURL string) (string, error) {
	statusCode, repoList, err := httpRequest(&fasthttp.Client{}, g.Version, fullURL, g.Token, g.UseCredentials)
	if err != nil {
		return "", errors.New("error while getting repository SPDX: " + strconv.Itoa(statusCode) + shared.Separator + err.Error())
	}
	if statusCode == http.StatusNotFound || statusCode == http.StatusMovedPermanently {
		log.Warnf("License not found for %s, skipping...", fullURL)
		return "", nil
	}

	var s spdx
	err = jsoniter.Unmarshal(repoList, &s)
	if err != nil {
		return "", errors.New("error while deserializing repository SPDX: " + strconv.Itoa(statusCode) + shared.Separator + err.Error())
	}

	if id, ok := s.License["spdx_id"]; ok {
		return id, nil
	}

	return "", nil
}

func httpRequest(client *fasthttp.Client, version, uri, token string, useCred bool) (statusCode int, body []byte, err error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(uri)
	req.Header.SetMethod(http.MethodGet)
	req.Header.Set(GithubVersionHeader, version)
	if useCred {
		req.Header.Set(Authorization, Bearer+token)
	}
	resp := fasthttp.AcquireResponse()
	err = client.Do(req, resp)
	if err != nil {
		return http.StatusInternalServerError, []byte{}, err
	}
	fasthttp.ReleaseRequest(req)

	body = resp.Body()
	statusCode = resp.StatusCode()

	fasthttp.ReleaseResponse(resp)
	return statusCode, body, nil
}
