package controller

import (
	"net/http"
	"scalingo/internal/core/domain"
	"scalingo/internal/core/port"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func ProvideRepoHTTPHandler(
	repoInterface port.RepoInterface,
) *RepoHTTPHandler {
	return &RepoHTTPHandler{
		repoInterface: repoInterface,
	}
}

type RepoHTTPHandler struct {
	repoInterface port.RepoInterface
}

func (p *RepoHTTPHandler) RepoController(ctx context.Context, c *gin.Context) {
	input, err := c.GetRawData()
	if err != nil {
		log.Errorf("List projects - unable to read input: %#v\n", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	domainInput := &domain.ListRepoInput{}

	err = jsoniter.Unmarshal(input, &domainInput)
	if err != nil {
		log.Errorf("List projects - unable to deserialize: %#v\n", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	err = validateListProjects(domainInput)
	if err != nil {
		log.Errorf("List projects - validation error: %#v\n", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	projectList, err := p.repoInterface.ListRepositories(ctx, domainInput)
	if err != nil {
		log.Errorf("List projects error: %#v\n", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projectList)
}
