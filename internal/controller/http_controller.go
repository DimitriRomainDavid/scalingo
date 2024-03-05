package controller

import (
	"net/http"
	"time"

	conf "scalingo/internal/infra/config"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func ProvideHTTPService(
	ctx context.Context,
	config *conf.Config,
	g *gin.Engine,
) *HTTPService {
	return &HTTPService{
		server: &http.Server{
			ReadHeaderTimeout: time.Second,
			Addr:              config.GetServerAddress(),
			Handler:           g,
		},
		config: config,
		ctx:    ctx,
	}
}

type HTTPService struct {
	server *http.Server
	config *conf.Config
	ctx    context.Context
}

func (h *HTTPService) StartHTTPServer() {
	go func() {
		if err := h.server.ListenAndServe(); err != nil {
			log.Error(err)
		}
	}()
}

func (h *HTTPService) ShutdownHTTPServer() {
	ctx, cancel := context.WithCancel(h.ctx)
	defer cancel()

	if err := h.server.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown err=%s", err.Error())
	}
}
