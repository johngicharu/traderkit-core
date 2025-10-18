package terminal

import (
	"backend/internal/common"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type accountType struct {
	Login  int
	Server string
}

type TerminalConnector struct {
	mu                sync.RWMutex
	terminalRawTcpUrl string
	accounts          map[string]accountType
	terminals         map[string]*Terminal
	ctx               context.Context
	cancel            context.CancelFunc
}

func NewTerminalConnector(terminalRawTcpUrl string) (*TerminalConnector, error) {
	return &TerminalConnector{
		accounts:          make(map[string]accountType),
		terminals:         make(map[string]*Terminal),
		terminalRawTcpUrl: terminalRawTcpUrl,
	}, nil
}

func (tc *TerminalConnector) Run(parentCtx context.Context) error {
	listener, err := net.Listen("tcp", tc.terminalRawTcpUrl)
	if err != nil {
		return fmt.Errorf("listen failed: %v", err)
	}

	defer listener.Close()
	log.Printf("[ctrl] terminal connector listening on: %s", tc.terminalRawTcpUrl)
	tc.ctx, tc.cancel = context.WithCancel(parentCtx)

	// want this one to block
	tc.acceptNewConnections(listener)

	return nil
}

func (tc *TerminalConnector) acceptNewConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			select {
			case <-tc.ctx.Done():
				return
			default:
				log.Printf("[ctrl] accept new terminal conn err: %v", err)
				time.Sleep(200 * time.Millisecond)
				continue
			}
		}

		// pass this new connection on to the respective terminal
		go tc.handleIncomingConnection(&conn)
	}
}

func (tc *TerminalConnector) handleIncomingConnection(conn *net.Conn) {
	remote := (*conn).RemoteAddr().String()

	terminal, err := tc.performHandshake(conn)
	if err != nil {
		log.Printf("[ctrl] handshake failed from %s: %v", remote, err)
		(*conn).Close()
		return
	}

	terminal.HandleConnection(tc.ctx, conn)
}

func (tc *TerminalConnector) performHandshake(conn *net.Conn) (term *Terminal, err error) {
	_ = (*conn).SetDeadline(time.Now().Add(200 * time.Millisecond))
	defer (*conn).SetDeadline(time.Time{})

	decoder := json.NewDecoder(*conn)
	var res common.TaskRes
	if err := decoder.Decode(&res); err != nil {
		return nil, err
	}

	if res.MiscDetails == nil {
		return nil, fmt.Errorf("[ctrl] invalid handshake msg")
	}

	if terminal, ok := tc.GetTerminal(res.MiscDetails.TerminalId); !ok {
		return nil, fmt.Errorf("[ctrl] terminal not found")
	} else {
		return terminal, nil
	}
}

func (tc *TerminalConnector) GetTerminal(id string) (terminal *Terminal, ok bool) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	terminal, ok = tc.terminals[id]
	return
}

// add functions for all account related actions

func (tc *TerminalConnector) AddAccount(id string, login int, server string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.accounts[id] = accountType{
		Login:  login,
		Server: server,
	}
}

func (tc *TerminalConnector) RemoveAccount(id string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	delete(tc.accounts, id)
}

func (tc *TerminalConnector) SendToTerminal(req common.TaskReq) error {
	if req.MiscDetails == nil {
		return fmt.Errorf("[ctrl] invalid request - missing terminal id")
	}

	termId := req.MiscDetails.TerminalId

	tc.mu.Lock()
	defer tc.mu.Unlock()

	if terminal, ok := tc.terminals[termId]; ok {
		terminal.SendTask(req)
	}

	return nil
}
