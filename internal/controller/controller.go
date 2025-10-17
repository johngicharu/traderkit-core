package controller

import (
	"backend/internal/common"
	servercomms "backend/internal/controller/server_comms"
	"backend/internal/controller/terminal"
	"context"
)

type Controller struct {
	TerminalComms *terminal.TerminalConnector
	ServerComms   *servercomms.ServerConnector
}

func NewController(ctx context.Context, conf common.ControllerConfig) (*Controller, error) {
	termComms, err := terminal.NewTerminalConnector()
	if err != nil {
		return nil, err
	}

	wsServer, err := servercomms.NewServerConnector(ctx, conf, termComms)
	if err != nil {
		return nil, err
	}

	return &Controller{
		TerminalComms: termComms,
		ServerComms:   wsServer,
	}, nil
}

func (ctrl *Controller) Run() error {
	return ctrl.ServerComms.Start()
}

/**
dispatcher
registry (list of terminals)
controller definition
*/
