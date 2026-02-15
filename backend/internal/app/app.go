package app

import (
	"context"
	"log"
	"net/http"

	"agenteur.ai/api/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type App struct {
	Config *config.Config
	Server *http.Server
	DB     *pgx.Conn
}

func NewApp() *App {
	config := config.Load()
	db, err := pgx.Connect(context.Background(), config.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(context.Background()); err != nil {
		log.Fatal("db ping failed:", err)
	}
	log.Println("db connected successfully")
	server := &http.Server{
		Addr:    config.Port,
		Handler: NewRouter(),
	}
	return &App{
		Config: config,
		Server: server,
		DB:     db,
	}
}

func (a *App) Start() error {
	log.Println("starting server on port", a.Config.Port)
	return a.Server.ListenAndServe()
}

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("server is awake"))
	})

	return r
}
