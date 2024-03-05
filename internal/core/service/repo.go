package service

import (
	"context"
	"scalingo/internal/core/domain"
	"scalingo/internal/core/dto"
	"scalingo/internal/core/port"
	conf "scalingo/internal/infra/config"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

func ProvideRepoService(config *conf.Config, github port.GithubInterface) *RepoService {
	return &RepoService{
		Config: config,
		Github: github,
	}
}

type RepoService struct {
	Config *conf.Config
	Github port.GithubInterface
}

// Concurrent safe repository ID holder for next batched request to GitHub
type lowestID struct {
	sync.RWMutex
	ID            int
	LastProcessed int
}

func (p *RepoService) ListRepositories(ctx context.Context, repoInput *domain.ListRepoInput) ([]*domain.ListRepoOutput, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	id, err := p.Github.GetLatestRepoID()
	if err != nil || id == 0 {
		return nil, err
	}

	lowestIDForNextBatch := &lowestID{ID: id, LastProcessed: id}
	listOutput := make([]*domain.ListRepoOutput, 0)
	requested := 0

	// While the output size if not fulfilled
	for requested < p.Config.OutputSize {
		reposChannel := make(chan *domain.ListRepoOutput, p.Config.ProcessingBatchSize)
		var currentList []*dto.LatestCreatedRepo
		currentList, err = p.Github.GetRepositories(id)
		if err != nil {
			return nil, err
		}
		log.Infof(
			"Current batch ID %d for %d number to retrieve and %d retrieved in last request, %d left",
			lowestIDForNextBatch.ID,
			p.Config.OutputSize,
			len(currentList),
			p.Config.OutputSize-len(currentList),
		)

		requested += len(currentList)

		var wg sync.WaitGroup
		wg.Add(len(currentList))

		// We iterate through the repos returned in the request and process them concurrently/in parallel
		for _, repository := range currentList {
			go func(
				ctx context.Context,
				repository *dto.LatestCreatedRepo,
			) { // loop var anonymous parameter not necessary anymore since 1.22, see https://go.dev/blog/loopvar-preview
				defer wg.Done()

				returnedRepository := &domain.ListRepoOutput{
					FullName:    repository.FullName,
					Owner:       repository.Owner.Login,
					Repository:  repository.HTMLURL,
					Description: repository.Description,
				}

				// Cancel context for leftover routines if the listOutput is full,
				// actually not really necessary since the above processing (returnedRepository assignation) is really fast
				// but for more intensives tasks it can be vital to free the resources as early as possible
				// The goal for this precise case would have been to reduce the number of requests to GitHub
				// as much as possible since there is a restrictive rate limit
				select {
				case <-ctx.Done():
					return
				default:
				}

				returnedRepository.Languages, err = p.Github.GetRepositoryLanguages(repository.LanguagesURL)
				if err != nil {
					log.Errorf("couldn't retrieve languages: %#v", err)
				}

				returnedRepository.License, err = p.Github.GetRepositorySPDX(repository.URL)
				if err != nil {
					log.Errorf("couldn't retrieve spdx: %#v", err)
				}

				// We want the lowest ID of the current batch to use it as the next 'since' query parameter to GitHub
				if lowestIDForNextBatch.ID < repository.ID {
					lowestIDForNextBatch.Lock()
					lowestIDForNextBatch.ID = repository.ID
					lowestIDForNextBatch.Unlock()
				}

				if p.filter(
					repoInput,
					repository,
					returnedRepository.License,
					returnedRepository.Languages,
					returnedRepository.RepoSize(),
				) {
					reposChannel <- returnedRepository
				}
			}(ctx, repository)
		}

		// Wait for the repos to be concurrently/parallel precessed and inserted in the channel
		// and then close the channel (non-blocking for the channel reads)
		go func() {
			wg.Wait()
			close(reposChannel)
		}()

		// Retrieve the repos from the channel of the current batch and assign a new id for the 'since' parameter of the next batch if necessary
		for pulledRepo := range reposChannel {
			listOutput = append(listOutput, pulledRepo)
			if len(listOutput) >= p.Config.OutputSize {
				break
			}
		}
		id = lowestIDForNextBatch.ID
	}

	return listOutput, nil
}

//nolint:gocyclo
func (p *RepoService) filter(
	repoInput *domain.ListRepoInput,
	repository *dto.LatestCreatedRepo,
	spdx string,
	languages map[string]int,
	repoSize int64,
) bool {
	validateFilters := map[string]bool{}

	if repoInput.NameContains != "" {
		switch strings.Contains(strings.ToLower(repository.Name), strings.ToLower(repoInput.NameContains)) {
		case true:
			validateFilters["name_contains"] = true
		case false:
			validateFilters["name_contains"] = false
		}
	}

	if repoInput.DescriptionContains != "" {
		switch strings.Contains(strings.ToLower(repository.Description), strings.ToLower(repoInput.DescriptionContains)) {
		case true:
			validateFilters["desc_contains"] = true
		case false:
			validateFilters["desc_contains"] = false
		}
	}

	if repoInput.MinSize > 0 {
		switch repoSize > repoInput.MinSize {
		case true:
			validateFilters["min_size"] = true
		case false:
			validateFilters["min_size"] = false
		}
	}

	if repoInput.MaxSize > 0 {
		switch repoSize < repoInput.MaxSize {
		case true:
			validateFilters["max_size"] = true
		case false:
			validateFilters["max_size"] = false
		}
	}

	if repoInput.Language != "" {
		for language := range languages {
			if strings.Contains(strings.ToLower(language), strings.ToLower(repoInput.Language)) {
				validateFilters["language"] = true
				break
			}
		}
		if _, ok := validateFilters["language"]; !ok {
			validateFilters["language"] = false
		}
	}

	if repoInput.License != "" {
		switch strings.Contains(strings.ToLower(spdx), strings.ToLower(repoInput.License)) {
		case true:
			validateFilters["spdx"] = true
		case false:
			validateFilters["spdx"] = false
		}
	}

	for _, filterStatus := range validateFilters {
		if !filterStatus {
			return false
		}
	}

	return true
}
