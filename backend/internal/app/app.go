package app

import (
	"log"
	"net/http"

	"agenteur.ai/api/internal/config"
)

type App struct {
	Config *config.Config
	Server *http.Server
}

func NewApp() *App {
	config := config.Load()
	server := &http.Server{
		Addr:    config.Port,
		Handler: NewRouter(),
	}
	return &App{
		Config: config,
		Server: server,
	}
}

func (a *App) Start() error {
	log.Println("Starting server on port", a.Config.Port)
	return a.Server.ListenAndServe()
}

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("server is awake"))
	})
	return mux
}
