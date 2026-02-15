package app

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"agenteur.ai/api/internal/config"
	"agenteur.ai/api/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type App struct {
	Config *config.Config
	Server *http.Server
	DB     *pgx.Conn
}

func NewApp() *App {
	cfg := config.Load()
	db, err := pgx.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(context.Background()); err != nil {
		log.Fatal("db ping failed:", err)
	}
	log.Println("db connected successfully")
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	).With(
		"service", "api",
		"env", cfg.Env,
	)
	server := &http.Server{
		Addr:    cfg.Port,
		Handler: NewRouter(cfg, logger),
	}
	return &App{
		Config: cfg,
		Server: server,
		DB:     db,
	}
}

func (a *App) Start() error {
	log.Println("starting server on port", a.Config.Port)
	return a.Server.ListenAndServe()
}

func NewRouter(cfg *config.Config, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler { return middleware.RequestID()(next) })
	r.Use(func(next http.Handler) http.Handler { return middleware.RequestLogger(logger)(next) })
	r.Use(func(next http.Handler) http.Handler { return middleware.CORS(cfg.CORSAllowedOrigins)(next) })

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("server is awake"))
	})

	r.Route("/api", func(api chi.Router) {
		// Future group-specific middleware (auth/admin) can be attached here.
	})

	return r
}
