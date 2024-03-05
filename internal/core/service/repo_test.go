package service

import (
	"context"
	"scalingo/internal/core/domain"
	"scalingo/internal/core/dto"
	"scalingo/internal/infra/config"
	"testing"

	"github.com/stretchr/testify/assert"
	s "github.com/stretchr/testify/suite"
)

type RepoServiceSuite struct {
	s.Suite
	repoService *RepoService
}

type mockGithub struct{}

func (m *mockGithub) GetLatestRepoID() (int, error) { return 1, nil }
func (m *mockGithub) GetRepositories(_ int) ([]*dto.LatestCreatedRepo, error) {
	return []*dto.LatestCreatedRepo{
		{
			ID:       1,
			Name:     "repo_one",
			FullName: "john_doe/repo_one",
			Owner: &dto.Owner{
				Login: "john_doe",
			},
			HTMLURL:      "https://github.com/john_doe/repo_one",
			LanguagesURL: "https://api.github.com/repos/john_doe/repo_one/languages",
			URL:          "https://api.github.com/repos/john_doe/repo_one",
			Description:  "first sample repository",
		},
		{
			ID:       2,
			Name:     "repo_two",
			FullName: "jane_doe/repo_two",
			Owner: &dto.Owner{
				Login: "jane_doe",
			},
			HTMLURL:      "https://github.com/jane_doe/repo_two",
			LanguagesURL: "https://api.github.com/repos/jane_doe/repo_two/languages",
			URL:          "https://api.github.com/repos/jane_doe/repo_two",
			Description:  "api sample repository",
		},
		{
			ID:       3,
			Name:     "repo_three",
			FullName: "alice_smith/repo_three",
			Owner: &dto.Owner{
				Login: "alice_smith",
			},
			HTMLURL:      "https://github.com/alice_smith/repo_three",
			LanguagesURL: "https://api.github.com/repos/alice_smith/repo_three/languages",
			URL:          "https://api.github.com/repos/alice_smith/repo_three",
			Description:  "third sample repository",
		},
		{
			ID:       4,
			Name:     "repo_four",
			FullName: "bob_jones/repo_four",
			Owner: &dto.Owner{
				Login: "bob_jones",
			},
			HTMLURL:      "https://github.com/bob_jones/repo_four",
			LanguagesURL: "https://api.github.com/repos/bob_jones/repo_four/languages",
			URL:          "https://api.github.com/repos/bob_jones/repo_four",
			Description:  "fourth sample repository",
		},
	}, nil
}

func (m *mockGithub) GetRepositoryLanguages(fullURL string) (map[string]int, error) {
	var languages map[string]int
	switch fullURL {
	case "https://api.github.com/repos/john_doe/repo_one/languages":
		languages = map[string]int{
			"Go":   10,
			"Java": 50,
		}
	case "https://api.github.com/repos/jane_doe/repo_two/languages":
		languages = map[string]int{
			"C#":   100,
			"Java": 50,
		}
	case "https://api.github.com/repos/alice_smith/repo_three/languages":
		languages = map[string]int{
			"Javascript": 1000,
		}
	case "https://api.github.com/repos/bob_jones/repo_four/languages":
		languages = map[string]int{
			"C++": 10000,
		}
	}
	return languages, nil
}

func (m *mockGithub) GetRepositorySPDX(fullURL string) (string, error) {
	var spdx string
	switch fullURL {
	case "https://api.github.com/repos/john_doe/repo_one":
		spdx = "MIT"
	case "https://api.github.com/repos/jane_doe/repo_two":
		spdx = "GPL-3.0"
	case "https://api.github.com/repos/alice_smith/repo_three":
		spdx = "AGPL-3.0"
	case "https://api.github.com/repos/bob_jones/repo_four":
		spdx = "BSD-2"
	}
	return spdx, nil
}

func (suite *RepoServiceSuite) SetupTest() {
	suite.repoService = ProvideRepoService(&config.Config{OutputSize: 4}, &mockGithub{})
}

func (suite *RepoServiceSuite) TearDownTest() {}

func (suite *RepoServiceSuite) TestListRepositories_NameContains() {
	repoInput := &domain.ListRepoInput{NameContains: "three"}
	output, err := suite.repoService.ListRepositories(context.Background(), repoInput)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), output)
	assert.Equal(suite.T(), 1, len(output))
	for _, repo := range output {
		assert.Contains(suite.T(), repo.FullName, "three")
	}
}

func (suite *RepoServiceSuite) TestListRepositories_DescriptionContains() {
	repoInput := &domain.ListRepoInput{DescriptionContains: "API"}
	output, err := suite.repoService.ListRepositories(context.Background(), repoInput)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), output)
	assert.Equal(suite.T(), 1, len(output))
	for _, repo := range output {
		assert.Contains(suite.T(), repo.Description, "api")
	}
}

func (suite *RepoServiceSuite) TestListRepositories_SizeRange() {
	repoInput := &domain.ListRepoInput{MinSize: 10, MaxSize: 500}
	output, err := suite.repoService.ListRepositories(context.Background(), repoInput)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), output)
	assert.Equal(suite.T(), 2, len(output))
	for _, repo := range output {
		size := repo.RepoSize()
		assert.GreaterOrEqual(suite.T(), size, int64(10))
		assert.LessOrEqual(suite.T(), size, int64(500))
	}
}

func (suite *RepoServiceSuite) TestListRepositories_Language() {
	repoInput := &domain.ListRepoInput{Language: "Go"}
	output, err := suite.repoService.ListRepositories(context.Background(), repoInput)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), output)
	assert.Equal(suite.T(), 1, len(output))
	for _, repo := range output {
		assert.Contains(suite.T(), repo.Languages, "Go")
	}
}

func (suite *RepoServiceSuite) TestListRepositories_License() {
	repoInput := &domain.ListRepoInput{License: "MIT"}
	output, err := suite.repoService.ListRepositories(context.Background(), repoInput)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), output)
	assert.Equal(suite.T(), 1, len(output))
	for _, repo := range output {
		assert.Equal(suite.T(), "MIT", repo.License)
	}
}

func TestRepoServiceSuite(t *testing.T) {
	s.Run(t, new(RepoServiceSuite))
}
