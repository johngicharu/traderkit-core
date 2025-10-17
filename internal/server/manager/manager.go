package manager

import (
	"backend/internal/common"
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

// tracks connected controllers
type Manager struct {
	registry *Registry
	upgrader websocket.Upgrader
	incoming chan common.TaskRes
	Outgoing chan common.TaskReq
	//reconnectChan  chan string
	//disconnectChan chan string
	ctx    context.Context
	cancel context.CancelFunc
}

func NewManager(parentCtx context.Context, registry *Registry) *Manager {
	ctx, cancel := context.WithCancel(parentCtx)

	return &Manager{
		registry: registry,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		incoming: make(chan common.TaskRes),
		Outgoing: make(chan common.TaskReq),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (m *Manager) Start() {
	log.Println("Started manager")
	for {
		select {
		case <-m.ctx.Done():
			log.Println("Ctx closed")
		case msg := <-m.Outgoing:
			log.Println(msg)
		}
	}
}

/*func (m *Manager) Start() {
	for {
		select {
		case id := <-m.reconnectChan:
			log.Printf("Controller %s connected/reconnected", id)

		case id := <-m.disconnectChan:
			log.Printf("Controller %s disconnected", id)
			//go m.retryConnect(id)
			m.registry.Delete(id)
		}
	}
}*/

func (m *Manager) HandleConnection(w http.ResponseWriter, r *http.Request) {
	/*token := r.URL.Query().Get("token")
	claims, err := ValidateToken(token)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}*/

	// Validation passed, upgrade
	id := r.Header.Get("X-Controller-Id")
	capacity, err := strconv.Atoi(r.Header.Get("X-Controller-Capacity"))
	if err != nil {
		log.Printf("Invalid capacity")
		return
	}

	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade failed: %v", err)
		return
	}

	c := m.registry.GetOrCreateController(id, capacity, m.incoming)
	c.SetConnection(m.ctx, conn)
}

func (m *Manager) Incoming() <-chan common.TaskRes {
	// consume these messages and prioritize trades from master accounts
	return m.incoming
}

/* Consume Messages from controllers -> and use the publishers if need be
go func() {
    for msg := range wsManager.Incoming() {
        // route to correct task handler, DB, etc.
        log.Printf("Processing message from %s: %s", msg.ControllerID, msg.Payload)
    }
}()
*/
