package server

import (
	"backend/internal/server/api"
	"backend/internal/server/manager"
	"context"
	"log"
	"net/http"
)

type Server struct {
	ApiServer    *http.Server
	CtrlsManager *manager.Manager
}

func NewServer(api_url string) (*Server, error) {
	registry := manager.NewRegistry()
	ctrl_manager := manager.NewManager(registry)

	app := &Server{
		ApiServer:    api.InitHandler(api_url, ctrl_manager),
		CtrlsManager: ctrl_manager,
	}

	return app, nil
}

func (a *Server) Run(parentCtx context.Context) error {
	go a.CtrlsManager.Start(parentCtx)

	log.Printf("[server] started server api")
	return a.ApiServer.ListenAndServe()
}
