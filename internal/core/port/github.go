package port

import "scalingo/internal/core/dto"

type GithubInterface interface {
	GetLatestRepoID() (int, error)
	GetRepositories(id int) ([]*dto.LatestCreatedRepo, error)
	GetRepositoryLanguages(fullURL string) (map[string]int, error)
	GetRepositorySPDX(fullURL string) (string, error)
}
