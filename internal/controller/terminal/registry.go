package terminal

import (
	"backend/internal/common"
	"context"
	"fmt"
	"sync"
)

type accountType struct {
	Login  int
	Server string
}

type TerminalConnector struct {
	mu        sync.RWMutex
	accounts  map[string]accountType
	terminals map[string]*Terminal
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewTerminalConnector() (*TerminalConnector, error) {
	return &TerminalConnector{
		accounts: make(map[string]accountType),
	}, nil
}

func (r *TerminalConnector) AddAccount(id string, login int, server string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.accounts[id] = accountType{
		Login:  login,
		Server: server,
	}
}

func (r *TerminalConnector) RemoveAccount(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.accounts, id)
}

func (r *TerminalConnector) HandleTerminalConn() {}

func (r *TerminalConnector) SendToTerminal(req common.TaskReq) error {
	if req.MiscDetails == nil {
		return fmt.Errorf("invalid request - missing terminal id")
	}

	termId := req.MiscDetails.TerminalId

	r.mu.Lock()
	defer r.mu.Unlock()

	if terminal, ok := r.terminals[termId]; ok {
		terminal.SendTask(req)
	}

	return nil
}
