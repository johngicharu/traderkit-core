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

func NewController(conf common.ControllerConfig) (*Controller, error) {
	termComms, err := terminal.NewTerminalConnector(conf.TerminalRawTcpUrl)
	if err != nil {
		return nil, err
	}

	wsServer, err := servercomms.NewServerConnector(conf, termComms)
	if err != nil {
		return nil, err
	}

	return &Controller{
		TerminalComms: termComms,
		ServerComms:   wsServer,
	}, nil
}

func (ctrl *Controller) Run(ctx context.Context) error {
	go ctrl.TerminalComms.Run(ctx)
	return ctrl.ServerComms.Start(ctx)
}

/**
dispatcher
registry (list of terminals)
controller definition
*/
