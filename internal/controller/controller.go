package controller

import (
	servercomms "backend/internal/controller/server_comms"
	"backend/internal/controller/terminal"
	"context"
)

type Controller struct {
	Registry    *terminal.Registry
	ServerComms *servercomms.ServerConnector
}

func NewController(ctx context.Context) (*Controller, error) {
	registry, err := terminal.NewRegistry()
	if err != nil {
		return nil, err
	}

	wsServer, err := servercomms.NewServerConnector(ctx, "", "", "", registry)
	if err != nil {
		return nil, err
	}

	return &Controller{
		Registry:    registry,
		ServerComms: wsServer,
	}, nil
}

func (a *Controller) Run() error {
	return nil
}
