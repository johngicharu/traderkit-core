package manager

import (
	"backend/internal/common"
	"fmt"
	"sync"
)

// keeps track of available controllers and their accounts

type Registry struct {
	mu           sync.RWMutex
	controllers  map[string]*Controller
	accountIndex map[string]string // AccountId -> controllerId
}

// TODO -> save these mappings to a store or cache them for restarts
func NewRegistry() *Registry {
	return &Registry{
		controllers:  make(map[string]*Controller),
		accountIndex: make(map[string]string),
	}
}

func (r *Registry) GetOrCreateController(id string, capacity int, outbound chan<- common.TaskRes) *Controller {
	r.mu.Lock()
	defer r.mu.Unlock()

	c := NewController(id, capacity, outbound)

	r.controllers[id] = c

	return c
}

func (r *Registry) Get(id string) (*Controller, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.controllers[id]

	return c, ok
}

/*
func (r *Registry) Update(id string, capacity int, connected bool) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if c, ok := r.controllers[id]; ok {
		c.Capacity = capacity
		c.Connected = connected
		c.UpdatedAt = time.Now()
	} else {
		return false
	}

	return true
}*/

func (r *Registry) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.controllers[id]; ok {
		delete(r.controllers, id)
	} else {
		return false
	}

	return true
}

// add setup details for the accounts mapping for master slave relationships
func (r *Registry) AssignAccount(controllerId, accId string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// maybe find a controller with capacity and send it there?
	// also, maybe base this on the region the controller is in

	if c, ok := r.controllers[controllerId]; ok {
		c.AppendAccount(accId)
		r.accountIndex[accId] = controllerId
	}
}

func (r *Registry) RemoveAccount(accId string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.accountIndex, accId)
	if c, ok := r.FindControllerByAccount(accId); ok {
		c.RemoveAccount(accId)
	}
}

func (r *Registry) FindControllerByAccount(accId string) (*Controller, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ctrlId, ok := r.accountIndex[accId]
	if !ok {
		return nil, false
	}

	c, ok := r.controllers[ctrlId]
	return c, ok
}

func (r *Registry) SendTaskToAccount(accountID string, payload []byte) error {
	ctrl, ok := r.FindControllerByAccount(accountID)
	if !ok {
		return fmt.Errorf("no controller for account %s", accountID)
	}

	if !ctrl.Connected() {
		return fmt.Errorf("controller %s not connected", ctrl.Id)
	}

	select {
	case ctrl.SendChan <- payload:
		return nil
	default:
		return fmt.Errorf("controller %s send buffer full", ctrl.Id)
	}
}
