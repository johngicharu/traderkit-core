package server

import (
	"backend/internal/server/api"
	"backend/internal/server/manager"
	"context"
	"net/http"
)

type Server struct {
	ApiServer    *http.Server
	CtrlsManager *manager.Manager
}

func NewServer(ctx context.Context, api_url string) (*Server, error) {
	registry := manager.NewRegistry()
	ctrl_manager := manager.NewManager(ctx, registry)

	app := &Server{
		ApiServer:    api.InitHandler(api_url, ctrl_manager),
		CtrlsManager: ctrl_manager,
	}

	return app, nil
}

func (a *Server) Run() error {
	go a.CtrlsManager.Start()

	return a.ApiServer.ListenAndServe()
}
