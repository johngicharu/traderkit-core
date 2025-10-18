package terminal

import (
	"backend/internal/common"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Terminal struct {
	mu       sync.RWMutex
	Id       string
	Type     common.TerminalType
	login    int
	server   string
	conn     net.Conn // raw tcp connection
	lastSeen time.Time

	taskRequests chan common.TaskReq

	// always update this
	taskResponse chan common.TaskRes
	ctx          context.Context
	cancel       context.CancelFunc
}

type TerminalDeploy struct {
	Id           string
	Type         common.TerminalType
	Login        int
	Password     string
	Server       string
	TradeAllowed bool
}

func NewTerminal(parentCtx context.Context, id string, termType common.TerminalType, login int, server string, tradeAllowed bool, conn net.Conn) *Terminal {
	return &Terminal{
		Id:           id,
		Type:         termType,
		login:        login,
		server:       server,
		conn:         conn,
		lastSeen:     time.Now(),
		taskRequests: make(chan common.TaskReq),
	}
}

func CreateTerminal(details TerminalDeploy) {
	// get login details from the user
	// get server file if present
	// pull mt4/5 image/data and have it present locally (maybe always have it ready and then just update when needed)
	// update config files inside mt4/5
	// create TerminalInstance (if user scheduled start when finished deploying, we just call start later)
}

func (term *Terminal) Start() {
	// start the terminal
	// maybe queue a task named "wait_for_terminal_startup" on the controller
	// controller waits for the terminal with that id to come online
	// if it doesn't within a specific time
	// try restarting the pod
	// queue another wait for terminal startup task
	// if it doesn't come online, delete the pod and start from create terminal again (of course pull terminal details - login, password, etc.)
	// log failure to admins and maybe push logs from mt4/5 to the admin log instance (send notification maybe)
}

func (term *Terminal) Restart() {
	// call stop
	// call start
}

func (term *Terminal) Stop() {
	// maybe separate normal and full close
	// so full close stops the ea first and then closes the chart (doesn't seem to be able to done in a single call no? -> maybe just send command to close the ea and then wait for response or no messages from terminal within 2 seconds, if none, we just mark it as stopped then return)
	// so a blocking execution (full stop)
	// once term is stopped, we close the pod and maybe even delete it since we will have to initialize using the config method
	// temporary/normal stop just shuts down the terminal so this is pod related
	// send task to ea to tell it to unload + close charts?
	// close the pod the terminal belongs to
	// send response of the task
}

func (term *Terminal) touch() {
	term.mu.Lock()
	defer term.mu.Unlock()
	term.lastSeen = time.Now()
}

func (term *Terminal) LastSeen() time.Time {
	term.mu.Lock()
	defer term.mu.Unlock()
	return term.lastSeen
}

func (term *Terminal) SendTask(task common.TaskReq) {
	term.mu.Lock()
	defer term.mu.Unlock()

	select {
	case <-term.ctx.Done():
		return
	case term.taskRequests <- task:
		return
	}
}

func (term *Terminal) UpdateResponseChan(termResChan chan common.TaskRes) {
	term.mu.Lock()
	defer term.mu.Unlock()
	term.taskResponse = termResChan
}

func (term *Terminal) HandleConnection(parentCtx context.Context, conn net.Conn) {
	term.mu.Lock()
	term.conn = conn
	term.ctx, term.cancel = context.WithCancel(parentCtx)
	term.lastSeen = time.Now()
	term.mu.Unlock()

	go term.taskResponseWriter()

	// this is blocking
	term.taskRequestReader()

	term.cancel()
	conn.Close()

	// unregister the terminal from the registry I believe
}

func (term *Terminal) taskRequestReader() {
	decoder := json.NewDecoder(term.conn)

	for {
		select {
		case <-term.ctx.Done():
			return
		default:
		}

		var msg common.TaskRes

		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				log.Printf("conn EOF %s", term.Id)
			} else {
				log.Printf("decode error from %s: %v", term.Id, err)
			}
		}

		// mark terminal as active
		term.touch()

		select {
		case term.taskResponse <- msg:
		default:
			log.Printf("inbound queue full, dropping message")
		}
	}
}

func (term *Terminal) taskResponseWriter() {
	for {
		select {
		case <-term.ctx.Done():
			return
		case data := <-term.taskRequests:
			rawBytes, err := json.Marshal(data)
			if err != nil {
				log.Printf("[term] failed to marshal req: %s: %v", term.Id, err)
				continue
			}

			var compactBuffer bytes.Buffer
			if err := json.Compact(&compactBuffer, rawBytes); err != nil {
				log.Printf("[term] error compacting buffer: %s: %v", term.Id, err)
				continue
			}

			if _, err := term.conn.Write(append(compactBuffer.Bytes(), '\n')); err != nil {
				log.Printf("[term] failed to send req to term: %s: %v", term.Id, err)
				continue
			}
		}
	}
}

func (term *Terminal) Shutdown() {
	term.mu.Lock()
	defer term.mu.Unlock()
	term.cancel()
}

func (term *Terminal) IdentifyTrades() {
	// find trades list from the terminal
	// check our db to confirm trade id's
	// if trade is found in our db, we link it's id
	// if not found, we issue a new id to the trade
	// let the main server map master-slave trade connections to allow for copying and editing
	// wait for terminal to send in new trade updates based on changes
	// terminal will cache it's list of trades and orders and if any change is detected, it should send an update for us to resync
}
