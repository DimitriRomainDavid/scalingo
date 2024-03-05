package main

import (
	"scalingo/internal/controller"

	"os"
	"os/signal"
	"syscall"
)

func ProvideApp(
	httpServer *controller.HTTPService,
) *App {
	return &App{
		httpServer: *httpServer,
	}
}

type App struct {
	httpServer controller.HTTPService
}

func (a *App) Start() {
	a.httpServer.StartHTTPServer()
	defer a.httpServer.ShutdownHTTPServer()

	c := make(chan os.Signal, 1)
	signal.Notify(
		c,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	<-c
}
