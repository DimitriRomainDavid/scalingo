package port

import (
	"golang.org/x/net/context"

	"scalingo/internal/core/domain"
)

type RepoInterface interface {
	ListRepositories(ctx context.Context, repoInput *domain.ListRepoInput) ([]*domain.ListRepoOutput, error)
}
