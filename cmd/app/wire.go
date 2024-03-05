//go:build wireinject

package main

import (
	"context"
	"github.com/google/wire"
	"scalingo/internal/controller"
	"scalingo/internal/core/port"
	"scalingo/internal/core/service"
	"scalingo/internal/infra/repositories"
	"scalingo/internal/router"

	"scalingo/internal/infra/config"
)

func InitializeApp() *App {
	wire.Build(
		ProvideApp,
		context.Background,
		config.ProvideConfig,

		router.ProvideRouter,
		controller.ProvideHTTPService,

		repositories.ProvideGithub,
		wire.Bind(new(port.GithubInterface), new(*repositories.Github)),

		controller.ProvideRepoHTTPHandler,
		service.ProvideRepoService,
		wire.Bind(new(port.RepoInterface), new(*service.RepoService)),
	)
	return &App{}
}
