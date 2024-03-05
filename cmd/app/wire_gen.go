// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"context"
	"scalingo/internal/controller"
	"scalingo/internal/core/service"
	"scalingo/internal/infra/config"
	"scalingo/internal/infra/repositories"
	"scalingo/internal/router"
)

// Injectors from wire.go:

func InitializeApp() *App {
	contextContext := context.Background()
	configConfig := config.ProvideConfig()
	github := repositories.ProvideGithub(configConfig)
	repoService := service.ProvideRepoService(configConfig, github)
	repoHTTPHandler := controller.ProvideRepoHTTPHandler(repoService)
	engine := router.ProvideRouter(contextContext, repoHTTPHandler, configConfig)
	httpService := controller.ProvideHTTPService(contextContext, configConfig, engine)
	app := ProvideApp(httpService)
	return app
}
