package common

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ControllerConfig struct {
	Id          string
	Token       string
	ServerWsUrl string
	Capacity    int
}

type TerminalType string

const (
	MT4 TerminalType = "mt4"
	MT5 TerminalType = "mt5"
)

func (t TerminalType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t *TerminalType) UnmarshalJSON(b []byte) error {
	var str string

	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}

	switch strings.TrimSpace(str) {
	case string(MT4):
		*t = MT4
	case string(MT5):
		*t = MT5
	default:
		return fmt.Errorf("invalid terminal type value: %s", str)
	}

	return nil
}

type TaskType string

const (
	AccountTask    TaskType = "account"
	TradeTask      TaskType = "trade"
	DataTask       TaskType = "data"
	ControllerTask TaskType = "controller"
	AckTask        TaskType = "ack"
)

func (t TaskType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t *TaskType) UnmarshalJSON(b []byte) error {
	var str string

	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}

	switch strings.TrimSpace(str) {
	case string(AccountTask):
		*t = AccountTask
	case string(TradeTask):
		*t = TradeTask
	case string(DataTask):
		*t = DataTask
	case string(ControllerTask):
		*t = ControllerTask
	case string(AckTask):
		*t = AckTask
	default:
		return fmt.Errorf("invalid task value: %s", str)
	}

	return nil
}

type TaskSubType string

const (
	AccountTaskDeploy   TaskSubType = "deploy"
	AccountTaskReDeploy TaskSubType = "redeploy"
	AccountTaskUnDeploy TaskSubType = "undeploy"
	AccountTaskDelete   TaskSubType = "delete"

	TradeTaskAdd TaskSubType = "add"
	TradeTaskMod TaskSubType = "modify"

	DataTaskSymbol    TaskSubType = "symbol"
	DataTaskPrice     TaskSubType = "price"
	DataTaskAccount   TaskSubType = "account"
	DataTaskTrades    TaskSubType = "trades"
	DataTaskPriceFeed TaskSubType = "price_feed"

	ControllerTaskShutdown          TaskSubType = "shutdown"
	ControllerTaskUpdateMt4Base     TaskSubType = "update_mt4"
	ControllerTaskUpdateMt5Base     TaskSubType = "update_mt5"
	ControllerTaskUpdateServerFiles TaskSubType = "update_server_files"
)

func (t TaskSubType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t *TaskSubType) UnmarshalJSON(b []byte) error {
	var str string

	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}

	switch strings.TrimSpace(str) {
	case string(AccountTaskDeploy):
		*t = AccountTaskDeploy
	case string(AccountTaskReDeploy):
		*t = AccountTaskReDeploy
	case string(AccountTaskUnDeploy):
		*t = AccountTaskUnDeploy
	case string(AccountTaskDelete):
		*t = AccountTaskDelete

	case string(TradeTaskAdd):
		*t = TradeTaskAdd
	case string(TradeTaskMod):
		*t = TradeTaskMod

	case string(DataTaskSymbol):
		*t = DataTaskSymbol
	case string(DataTaskPrice):
		*t = DataTaskPrice
	case string(DataTaskAccount):
		*t = DataTaskAccount
	case string(DataTaskTrades):
		*t = DataTaskTrades
	case string(DataTaskPriceFeed):
		*t = DataTaskPriceFeed

	case string(ControllerTaskShutdown):
		*t = ControllerTaskShutdown
	case string(ControllerTaskUpdateMt4Base):
		*t = ControllerTaskUpdateMt4Base
	case string(ControllerTaskUpdateMt5Base):
		*t = ControllerTaskUpdateMt5Base
	case string(ControllerTaskUpdateServerFiles):
		*t = ControllerTaskUpdateServerFiles

	default:
		return fmt.Errorf("invalid task value: %s", str)
	}

	return nil
}

/*
No need for execution subtypes
-> if volume is increasing for a ticket 0 trade, we add a trade
-> if volume is reducing for a ticket defined trade, we partially close
-> if volume is changing to 0 for a ticket defined trade, we fully close
-> if ticket is not defined, it is considered a new trade, unless vol is 0
*/

type DeployReq struct {
	AccountId  int    `json:"account_login"`
	Server     string `json:"server"`
	Password   string `json:"password"`
	ServerFile string `json:"server_file"`
}

// deploy_ea task (run on a separate process -> doesn't allow dll's)

// sent by users, trade copy, etc. -> used to fetch data or req executions
type TerminalMiscData struct {
	AccountId     int    `json:"account_login"`
	Server        string `json:"server"`
	IntervalSec   int    `json:"interval_seconds,omitempty"`
	IntervalStart int    `json:"interval_start,omitempty"`
	IntervalEnd   int    `json:"interval_end,omitempty"`
}

type TaskReq struct {
	Id          int               `json:"request_id"`
	ReqType     TaskType          `json:"type"`
	ReqSubType  TaskSubType       `json:"sub_type,omitempty"`
	MiscDetails *TerminalMiscData `json:"misc,omitempty"`
	// parsed based on the type of message
	Payload []byte `json:"payload,omitempty"`
}

type SymbolTaskPayload struct {
	Symbol    string `json:"symbol"`
	Timeframe string `json:"timeframe"`
	Shift     int    `json:"shift"`
}

type TradeExecTaskPayload struct {
	Ticket     int    `json:"ticket,omitempty"`
	Volume     int    `json:"volume"`
	EntryPrice int    `json:"entry_price,omitempty"`
	Expiry     int    `json:"expiry,omitempty"`
	Magic      int    `json:"magic,omitempty"`
	StopLoss   int    `json:"stop_loss,omitempty"`
	TakeProfit int    `json:"take_profit,omitempty"`
	Slippage   int    `json:"slippage,omitempty"`
	Symbol     string `json:"symbol"`
	OrderType  string `json:"order_type"`
	Comment    string `json:"comment"`
}

type TaskRes struct {
	ReqId       int               `json:"request_id"`
	ReqType     string            `json:"type"`
	ReqSubType  string            `json:"sub_type"`
	MiscDetails *TerminalMiscData `json:"misc,omitempty"`
	Err         string            `json:"error"`
	Payload     []byte            `json:"payload"`
}

func (tr *TaskRes) TradeResToTradeReq() (req *TradeExecTaskPayload, err error) {
	req = nil

	if err = json.Unmarshal(tr.Payload, req); err != nil {
		return nil, err
	}

	return req, nil
}
