package controller

import (
	"backend/internal/common"
	servercomms "backend/internal/controller/server_comms"
	"backend/internal/controller/terminal"
	"context"
)

type Controller struct {
	Registry    *terminal.Registry
	ServerComms *servercomms.ServerConnector
}

func NewController(ctx context.Context, conf common.ControllerConfig) (*Controller, error) {
	registry, err := terminal.NewRegistry()
	if err != nil {
		return nil, err
	}

	wsServer, err := servercomms.NewServerConnector(ctx, conf, registry)
	if err != nil {
		return nil, err
	}

	return &Controller{
		Registry:    registry,
		ServerComms: wsServer,
	}, nil
}

func (ctrl *Controller) Run() error {
	return ctrl.ServerComms.Start()
}
