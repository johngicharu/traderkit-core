package servercomms

import (
	"backend/internal/common"
	"backend/internal/controller/terminal"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ServerConnector struct {
	serverUrl    string
	authToken    string
	controllerId string
	capacity     int

	conn          *websocket.Conn
	dialer        *websocket.Dialer
	registry      *terminal.TerminalConnector // registry of present accounts
	taskRequests  chan common.TaskReq
	taskResponses chan common.TaskRes
	registered    bool // whether we informed the server or not that we are live

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

func NewServerConnector(conf common.ControllerConfig, registry *terminal.TerminalConnector) (*ServerConnector, error) {
	return &ServerConnector{
		serverUrl:    conf.ServerWsUrl,
		authToken:    conf.Token,
		controllerId: conf.Id,
		capacity:     conf.Capacity,

		dialer: &websocket.Dialer{HandshakeTimeout: 5 * time.Second},

		registry:      registry,
		taskRequests:  make(chan common.TaskReq),
		taskResponses: make(chan common.TaskRes),
	}, nil
}

func (sc *ServerConnector) Start(parentCtx context.Context) error {
	sc.ctx, sc.cancel = context.WithCancel(parentCtx)

	log.Println("[ctrl] starting controller")
	for {
		select {
		case <-sc.ctx.Done():
			return nil
		default:
			if err := sc.connect(); err != nil {
				log.Printf("[ctrl-ws] connect error: %v, retrying in 5s", err)
				time.Sleep(5 * time.Second)
				continue
			}
		}

		// blocks until the connection is closed and it exits to try to connect again
		sc.handleConnection()
		log.Println("[ctrl-ws] disconnected retrying in 5s...")
		time.Sleep(5 * time.Second)
	}
}

func (sc *ServerConnector) connect() error {
	headers := http.Header{}
	headers.Add("X-Controller-Id", sc.controllerId)
	headers.Add("X-Controller-Capacity", strconv.Itoa(sc.capacity))
	if sc.authToken != "" {
		headers.Add("Authorization", "Bearer "+sc.authToken)
	}

	conn, resp, err := sc.dialer.Dial(sc.serverUrl, headers)
	if err != nil {
		if resp != nil {
			log.Printf("[ctrl-ws] handshake failed with status: %d", resp.StatusCode)
		}

		return err
	}

	sc.mu.Lock()
	sc.conn = conn
	sc.registered = true
	sc.mu.Unlock()

	log.Println("[ctrl-ws] connected to ", sc.serverUrl)
	return nil
}

func (sc *ServerConnector) handleConnection() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		conn := func() *websocket.Conn {
			sc.mu.RLock()
			defer sc.mu.RUnlock()
			return sc.conn
		}()

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				log.Printf("[ctrl-ws] read error: %v", err)
				sc.closeConn()
				return
			}

			var task common.TaskReq
			if err := json.Unmarshal(data, &task); err != nil {
				log.Printf("[ctrl-ws] invalid task: %v", err)
				continue
			}

			/**
							send to the respective channel so for instance have terminal requests
			controller requests, etc. So each can have a goroutine listening to it's type of messages. For terminal related ones, they can be directly sent to the right terminal instead of routing it separately or having that entire goroutine
						**/
			select {
			case sc.taskRequests <- task:
				go sc.sendAck(task.Id)
			case <-sc.ctx.Done():
				return
			}

			// send server ack
		}
	}()

	// writer
	go func() {
		defer wg.Done()

		for {
			select {
			case <-sc.ctx.Done():
				return
			case taskRes := <-sc.taskResponses:
				data, _ := json.Marshal(taskRes)
				sc.mu.RLock()
				c := sc.conn
				sc.mu.RUnlock()

				if c == nil {
					log.Printf("[ctrl-ws] write failed no active connection")
					sc.queueFailedResponse(taskRes)
					return
				}

				if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
					log.Printf("[ctrl-ws] write error: %v", err)
					sc.closeConn()
					return
				}
			}
		}
	}()

	wg.Wait()
}

func (sc *ServerConnector) sendAck(taskID int) {
	// we never ack the server messages, rather, we resend responses from the terminals if they are not acked
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	if sc.conn == nil {
		return
	}

	ack := map[string]any{"type": "ack", "taskId": taskID}
	data, _ := json.Marshal(ack)
	if err := sc.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("[ctrl-ws] ack write error: %v", err)
	}
}

func (sc *ServerConnector) closeConn() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.conn != nil {
		sc.conn.Close()
		sc.conn = nil
	}

	sc.registered = false
}

func (sc *ServerConnector) queueFailedResponse(res common.TaskRes) {
	// maybe store this in redis or a db somewhere
	// on reconnect, we slowly pass them on to the main server or queue them
	select {
	case sc.taskResponses <- res:
		// queued again for resend
	default:
		log.Println("[ctrl-ws] dropping response, queue full")
	}
}

func (sc *ServerConnector) Stop() {
	sc.cancel()
	sc.closeConn()
}
