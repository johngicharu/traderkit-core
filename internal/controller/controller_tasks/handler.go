package controllertasks

import (
	"backend/internal/common"
	"backend/internal/controller/terminal"
)

type TaskHandler struct {
	registry *terminal.TerminalConnector
}

func NewTaskHandler(registry *terminal.TerminalConnector) *TaskHandler {
	return &TaskHandler{
		registry: registry,
	}
}

func (th *TaskHandler) handleTaskRequest(task common.TaskReq) {
	if task.ReqType == common.AckTask {
		// delete the file
	} else if _, ok := common.TerminalTasks[task.ReqType]; ok {
		// handle terminal related tasks
		th.registry.SendToTerminal(task)
	} else if task.ReqType == common.AccountTask {
		// handle the account related tasks
	} else {
		// handle controller related tasks
	}
}
