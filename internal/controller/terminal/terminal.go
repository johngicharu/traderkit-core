package terminal

import (
	"backend/internal/common"
	"backend/internal/common/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PodmanDetails struct {
	volumePath string
	configPath string
	execPath   string
	podId      string
	createdAt  time.Time
}

type Terminal struct {
	mu             sync.RWMutex
	Id             string
	Type           common.TerminalType
	login          int
	server         string
	conn           *net.Conn // raw tcp connection
	lastSeen       time.Time
	terminalPath   string
	tradingAllowed bool
	pod            *PodmanDetails

	taskRequests chan common.TaskReq

	// always update this
	taskResponse chan common.TaskRes
	ctx          context.Context
	cancel       context.CancelFunc
}

type TerminalDeploy struct {
	Id             string              `json:"id"`
	Type           common.TerminalType `json:"type"`
	Login          int                 `json:"login"`
	Password       string              `json:"password"`
	Broker         string              `json:"broker"`
	Server         string              `json:"server"`
	ServerFile     string              `json:"server_file"`
	TradingAllowed bool                `json:"trading_allowed"`
}

func NewTerminal(parentCtx context.Context, id string, termType common.TerminalType, login int, server string, terminalPath string, tradingAllowed bool, conn *net.Conn) *Terminal {
	return &Terminal{
		Id:             id,
		Type:           termType,
		login:          login,
		server:         server,
		conn:           conn,
		terminalPath:   terminalPath,
		tradingAllowed: tradingAllowed,
		lastSeen:       time.Now(),
		taskRequests:   make(chan common.TaskReq),
	}
}

// remember to clear the directories if a failure occurs
func setupAndStartPodmanContainer(details TerminalDeploy) (*PodmanDetails, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	terminalDir := filepath.Join(homeDir, "terminals", details.Id)

	// TODO -> copy files to the created directory from our default directories

	if err := os.MkdirAll(terminalDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("[ctrl] failed to create terminal directories: %v", err)
	}

	configPath := filepath.Join(terminalDir, "config.ini")
	execPath := filepath.Join(terminalDir, "terminal64.exe")
	if details.Type == common.MT4 {
		execPath = filepath.Join(terminalDir, "terminal.exe")
	}

	if err := replaceConfigDetails(configPath, details.Login, details.Password, details.Server, details.Type); err != nil {
		return nil, err
	}

	// create the pod container
	// use the podman create command to only create and link volumes but do not start

	return &PodmanDetails{
		volumePath: terminalDir,
		configPath: configPath,
		execPath:   execPath,
		createdAt:  time.Now(), // if we don't have any lastSeen greater than this, we need to delete this pod or at least recreate it -> cleanup task basically
	}, nil
}

func replaceConfigDetails(config_path string, login int, password string, server string, term_type common.TerminalType) error {
	// these files will not be packaged, so it's better to have them as strings somewhere
	mainConfig := config.MT5Config

	if term_type == common.MT4 {
		mainConfig = config.MT4Config
	}

	updatedWithLogin := strings.Replace(mainConfig, "{{login}}", strconv.Itoa(login), 1)
	updatedWithPassword := strings.Replace(updatedWithLogin, "{{password}}", password, 1)
	updatedWithServer := strings.Replace(updatedWithPassword, "{{server}}", server, 1)

	// should not attempt to create since we should have the dir already
	/*file, err := os.OpenFile(config_path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("failed to open or create: ", err)
		return err
	}

	file.Close()*/

	err := os.WriteFile(config_path, []byte(updatedWithServer), 0644)
	if err != nil {
		return err
	}

	return nil
}

func CreateTerminal(details TerminalDeploy) error {
	// get login details from the user
	// get server file if present
	if details.ServerFile == "" {
		// fetch the correct broker file from our servers
	}

	// pull mt4/5 image/data and have it present locally (maybe always have it ready and then just update when needed)
	podDetails, err := setupAndStartPodmanContainer(details)
	if err != nil {
		return err
	}

	if os.Getenv("USE_PODS") == "true" {
		if err := runTerminalPodman(podDetails); err != nil {
			return err
		}
	} else {
		if err := runTerminalRaw(podDetails); err != nil {
			return err
		}
	}

	// create TerminalInstance (if user scheduled start when finished deploying, we just call start later)

	return nil
}

func runTerminalRaw(pod *PodmanDetails) error {
	if _, err := os.Stat(pod.execPath); os.IsNotExist(err) {
		return fmt.Errorf("executable not found: %s", pod.execPath)
	}

	cmd := exec.Command(pod.execPath, "/portable", pod.configPath)
	cmd.Dir = filepath.Dir(pod.execPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Launching portable terminal: %s\n", pod.execPath)
	return cmd.Start() // use Wait() if you want blocking
}

func runTerminalPodman(pod *PodmanDetails) error {
	log.Printf("podman terminal run not implemented: %s", pod.podId)
	return nil
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

func (term *Terminal) HandleConnection(parentCtx context.Context, conn *net.Conn) {
	if conn == nil {
		return
	}

	term.mu.Lock()
	term.conn = conn
	term.ctx, term.cancel = context.WithCancel(parentCtx)
	term.lastSeen = time.Now()
	term.mu.Unlock()

	go term.taskResponseWriter()

	// this is blocking
	term.taskRequestReader()

	term.cancel()
	(*conn).Close()

	// unregister the terminal from the registry I believe
}

func (term *Terminal) taskRequestReader() {
	decoder := json.NewDecoder(*term.conn)

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
			if term.conn == nil {
				term.Shutdown()
				return
			}

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

			if _, err := (*term.conn).Write(append(compactBuffer.Bytes(), '\n')); err != nil {
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
