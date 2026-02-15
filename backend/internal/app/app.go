package app

import (
	"log"
	"net/http"

	"agenteur.ai/api/internal/config"

	"github.com/go-chi/chi/v5"
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
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("server is awake"))
	})

	return r
}
