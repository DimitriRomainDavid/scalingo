package router

import (
	"golang.org/x/net/context"

	"scalingo/internal/controller"
	conf "scalingo/internal/infra/config"

	"github.com/gin-gonic/gin"
)

func ProvideRouter(ctx context.Context, repositoriesController *controller.RepoHTTPHandler, config *conf.Config) *gin.Engine {
	gin.SetMode(config.GinMode)
	g := gin.Default()

	g.GET("/repositories", func(c *gin.Context) { repositoriesController.RepoController(ctx, c) })
	return g
}
