package terminal

import (
	"backend/internal/common"
	"sync"
	"time"
)

type Terminal struct {
	Id               string
	Type             common.TerminalType
	Login            int
	Password         string
	InvestorPassword string
	Server           string
	Broker           string
	BrokerId         string // used to name the broker files
	TradeAllowed     bool
	UpdatedAt        time.Time
}

type accountType struct {
	Login  int
	Server string
}

type Registry struct {
	mu       sync.RWMutex
	accounts map[string]accountType
}

func NewRegistry() (*Registry, error) {
	return &Registry{
		accounts: make(map[string]accountType),
	}, nil
}

func (r *Registry) AddAccount(id string, login int, server string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.accounts[id] = accountType{
		Login:  login,
		Server: server,
	}
}

func (r *Registry) RemoveAccount(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.accounts, id)
}
