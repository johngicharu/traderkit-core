package accounts

import (
	"backend/internal/common"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// have an accountId -> controller map
// controller will have accountId -> accountLogin + server map to correctly send to the right account

type AccountsApiService struct {
	outgoingChan chan common.TaskReq
}

func NewAccountsApiService(ctrlChan chan common.TaskReq) *AccountsApiService {
	return &AccountsApiService{
		outgoingChan: ctrlChan,
	}
}

func (a *AccountsApiService) deployAccount(w http.ResponseWriter, r *http.Request) {
	var account struct {
		Login             int    `json:"login"`
		Server            string `json:"server"`
		Password          string `json:"password"`
		Type              string `json:"type"`
		Investor_password string `json:"investor_password,omitempty"`
		Server_file       string `json:"server_file,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		fmt.Println(err)
		http.Error(w, "invalid data", http.StatusBadRequest)
		return
	}

	payloadBytes, err := json.Marshal(account)
	if err != nil {
		log.Printf("error marshaling account: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	select {
	case a.outgoingChan <- common.TaskReq{
		Id: 394,
		MiscDetails: &common.TerminalMiscData{
			AccountId: account.Login,
			Server:    account.Server,
		},
		ReqType:    common.AccountTask,
		ReqSubType: common.AccountTaskDeploy,
		Payload:    payloadBytes,
	}:
		w.Write([]byte("succesfully submitted account deployment task. please wait for updates"))
	default:
		http.Error(w, "failed to queue deployment task. please try again later", http.StatusRequestTimeout)
	}

	w.Write([]byte("succesfully deployed account"))
}

/*
	func (a *AccountsApiService) reDeployAccount(w http.ResponseWriter, r *http.Request) {
		// kind of equivalent to a restart or reinitialize

		w.Write([]byte("redeploy for account: " + r.PathValue("account_id")))
	}
*/
func (a *AccountsApiService) deleteAccount(w http.ResponseWriter, r *http.Request) {
	// delete from the db -> find and delete (find the correct id and server for the controller)
	server := ""
	accountId := 93480024

	// might need a periodic check of the db and the controllers to make sure no residual accounts remain

	// send task to the controller - based on the id
	select {
	case a.outgoingChan <- common.TaskReq{
		Id: 987384,
		MiscDetails: &common.TerminalMiscData{
			AccountId: accountId,
			Server:    server,
		},
		ReqType:    common.AccountTask,
		ReqSubType: common.AccountTaskDelete,
	}:
		w.Write([]byte("successfully queued delete action"))
	default:
		http.Error(w, "failed to queue account deletion task", http.StatusRequestTimeout)
	}
	w.Write([]byte("delete account for account: " + r.PathValue("account_id")))
}
func (a *AccountsApiService) shutdownAccount(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("shutdown for account: " + r.PathValue("account_id")))
}

// change details
func (a *AccountsApiService) updateAccount(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("update for account: " + r.PathValue("account_id")))
}

func (a *AccountsApiService) messageAccount(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("msg received for acc: " + r.PathValue("account_id")))

	req := common.TaskReq{Id: 89394}

	select {
	case a.outgoingChan <- req:
		return
	default:
		log.Println("failed to send msg to channel")
		return
	}

}

func (a *AccountsApiService) ApiHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /{account_id}/deploy", a.deployAccount)
	// mux.HandleFunc("POST /{account_id}/redeploy", a.reDeployAccount)
	mux.HandleFunc("POST /{account_id}/message", a.messageAccount)
	mux.HandleFunc("DELETE /{account_id}/delete", a.deleteAccount)
	mux.HandleFunc("PATCH /{account_id}/shutdown", a.shutdownAccount)
	mux.HandleFunc("PATCH /{account_id}/update", a.updateAccount)

	return mux
}
