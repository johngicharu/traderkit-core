package api

import (
	"backend/internal/server/api/users"
	"backend/internal/server/manager"
	"net/http"
	// cors "github.com/jub0bs/cors"
)

// contains REST endpoints for user commands, account linking, etc.
func apiHandler(ctrlManager *manager.Manager) http.Handler {

	/*corsMw, err := cors.NewMiddleware(cors.Config{
		Origins: []string{
			"*",
		},
	})

	if err != nil {
		log.Fatal(err)
	}*/

	mux := http.NewServeMux()

	usersService := &users.UsersApiService{}
	// mux.Handle("/api/users", http.StripPrefix("/api/users", corsMw.Wrap(usersService.ApiHandler(ctrlManager.Outgoing))))

	mux.Handle("/api/users/", http.StripPrefix("/api/users", usersService.ApiHandler(ctrlManager.Outgoing)))
	mux.HandleFunc("/ws", ctrlManager.HandleConnection)
	mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test endpoint hit"))
	})

	return mux
}

func InitHandler(endpoint string, ctrlManager *manager.Manager) *http.Server {
	apiServer := &http.Server{
		Addr:    endpoint,
		Handler: apiHandler(ctrlManager),
	}

	return apiServer
}
