package users

import (
	"backend/internal/common"
	"backend/internal/server/api/users/accounts"
	"net/http"
)

type UsersApiService struct {
}

func (u *UsersApiService) register(w http.ResponseWriter, r *http.Request)      {}
func (u *UsersApiService) login(w http.ResponseWriter, r *http.Request)         {}
func (u *UsersApiService) update(w http.ResponseWriter, r *http.Request)        {}
func (u *UsersApiService) subscribe(w http.ResponseWriter, r *http.Request)     {}
func (u *UsersApiService) unsubscribe(w http.ResponseWriter, r *http.Request)   {}
func (u *UsersApiService) deleteDetails(w http.ResponseWriter, r *http.Request) {}

func (u *UsersApiService) ApiHandler(ctrlChan chan common.TaskReq) http.Handler {
	mux := http.NewServeMux()

	accService := accounts.NewAccountsApiService(ctrlChan)

	mux.Handle("/accounts/", http.StripPrefix("/accounts", accService.ApiHandler()))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hit users endpoint"))
	})

	return mux
}
