package controllers

import "net/http"

// crud related to controller servers -> 1 controller = 1 controller server
/* maybe include configs for controllers here
- like fetch the details of a controller
- fetch the accounts the controller would use
- fetch the config for mt terminals
*/

// admin-level requests
func AuthController(w http.ResponseWriter, r *http.Request) {
	// controller sends their auth credentials and receives back an auth token to be used in the manager
}

func SendControllerMsg(w http.ResponseWriter, r *http.Request) {
	// send commands such as shutdown, update, restart -> related to the controller and not the account
}

/**
Summaries related to controllers, their accounts, etc. will be gotten directly from the DB

*/
