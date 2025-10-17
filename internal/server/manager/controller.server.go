package manager

import (
	"backend/internal/common"
	"context"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// defines controller metadata
type Controller struct {
	// protects fields such as capacity, length, accounts, connected state, updatedAt
	mu sync.RWMutex
	Id string
	// IP        string
	capacity  int
	length    int
	connected bool
	accounts  map[string]struct{} // account id's
	updatedAt time.Time

	// communication
	conn     *websocket.Conn
	SendChan chan []byte // outgoing messages to the controller
	outbound chan<- common.TaskRes

	ctx    context.Context
	cancel context.CancelFunc
}

/**
Before we handle accounts, we need to
Init a new controller instance
Pass the connection to it so it can handle the read and write loops
When the parent context is closed, it should also close
If it encounters an error while reading, it should close the connection and wait for a reconnection by the controller server
Registry doesn't need any locks for handling controller connect checks since it can be exposed via a function I think
Same applies to the number of accounts connected (length vs capacity)
*/

func NewController(id string, capacity int, outbound chan<- common.TaskRes) *Controller {
	return &Controller{
		Id:        id,
		capacity:  capacity,
		updatedAt: time.Now(),
		SendChan:  make(chan []byte),
		accounts:  make(map[string]struct{}),
		length:    0,
		connected: false,
		conn:      nil,
		outbound:  outbound,
	}
}

func (c *Controller) Connected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

func (c *Controller) SetConnection(parentCtx context.Context, conn *websocket.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		log.Printf("[%s] Closing old connection", c.Id)
		c.cancel() // cancel old context
		// c.wg.Wait()      // wait for all goroutines to finish
		if c.conn != nil {
			c.conn.Close()
		}
	}

	c.ctx, c.cancel = context.WithCancel(parentCtx)

	c.conn = conn
	c.connected = true
	c.updatedAt = time.Now()

	go c.readLoop()
	go c.writeLoop()
}

func (c *Controller) AppendAccount(accId string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// check for whether we are full already before appending
	c.accounts[accId] = struct{}{}
	c.length = len(c.accounts)
	c.updatedAt = time.Now()
	// update db here
}

func (c *Controller) RemoveAccount(accId string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.accounts, accId)
	c.updatedAt = time.Now()
	// update db call here
}

func (c *Controller) readLoop() {
	// defer c.wg.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, msg, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("[%s] Read error: %v", c.Id, err)
				// c.outbound <- ManagerMessage{ControllerId: c.Id, Type: "disconnect"}
				c.cancel()
				return
			}

			log.Printf("[%s] msg: %s", c.Id, msg)
			// c.outbound <- ManagerMessage{ControllerId: c.Id, Type: "message", Data: msg}
		}
	}
}

func (c *Controller) writeLoop() {
	// defer c.wg.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.SendChan:
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("[%s] Write error: %v", c.Id, err)
				// c.outbound <- ManagerMessage{ControllerId: c.Id, Type: "disconnect"}
				c.cancel()
				return
			}
		}
	}
}

/*func (c *Controller) Disconnect() {
    if c.cancel != nil {
        c.cancel()
    }
    c.wg.Wait()
    if c.Conn != nil {
        _ = c.Conn.Close()
    }
    c.Connected = false
    // Notify manager that we’re closed
    c.OutChan <- ControllerMessage{ControllerID: c.Id, Event: "closed"}
}*/

/**
Tasks
type account
sub_type deploy redeploy, delete, undeploy

type trade
subtype add, modify

type data
subtype symbol, price, account, trades, price_feed

default tasks (if master -> if slave, we do nothing)
type data -> trades

on frontend load -> send the most recent update of the user
	**/

// used for messaging between controllers such as ack, deployment, etc.
type InternalMsg struct {
}
