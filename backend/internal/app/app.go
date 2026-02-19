package app

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	adminhandlers "agenteur.ai/api/internal/administration/handlers"
	adminservices "agenteur.ai/api/internal/administration/services"
	authhandlers "agenteur.ai/api/internal/auth/handlers"
	authservices "agenteur.ai/api/internal/auth/services"
	"agenteur.ai/api/internal/config"
	"agenteur.ai/api/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Config *config.Config
	Server *http.Server
	DB     *pgxpool.Pool
}

func NewApp() *App {
	cfg := config.Load()
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("db ping failed:", err)
	}
	log.Println("db connected successfully")
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	).With(
		"service", "api",
		"env", cfg.Env,
	)

	// Auth domain
	userRepo := authservices.NewUserRepository()
	tokenRepo := authservices.NewRefreshTokenRepository()
	authService := authservices.NewAuthService(pool, userRepo, tokenRepo, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL, cfg.BcryptCost)
	userService := authservices.NewUserService(pool, userRepo)
	authMiddleware := authhandlers.NewAuthMiddleware(cfg.JWTSecret)
	secureCookies := cfg.Env != "local"
	authHandler := authhandlers.NewAuthHandler(authService, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL, secureCookies)
	userHandler := authhandlers.NewUserHandler(userService)

	// Administration domain
	orgRepo := adminservices.NewOrganizationRepository()
	membershipRepo := adminservices.NewMembershipRepository()
	orgService := adminservices.NewOrgService(pool, orgRepo, membershipRepo)
	orgHandler := adminhandlers.NewOrgHandler(orgService)
	roleMW := adminhandlers.NewRoleMiddleware(pool, membershipRepo, userRepo)

	server := &http.Server{
		Addr:    cfg.Port,
		Handler: NewRouter(cfg, logger, authMiddleware, roleMW, authHandler, userHandler, orgHandler),
	}
	return &App{
		Config: cfg,
		Server: server,
		DB:     pool,
	}
}

func (a *App) Start() error {
	log.Println("starting server on port", a.Config.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-quit:
		log.Println("shutting down gracefully...")
		a.DB.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 10*1e9) // 10s
		defer cancel()
		return a.Server.Shutdown(ctx)
	}
}

func NewRouter(
	cfg *config.Config,
	logger *slog.Logger,
	authMiddleware *authhandlers.AuthMiddleware,
	roleMW *adminhandlers.RoleMiddleware,
	authHandler *authhandlers.AuthHandler,
	userHandler *authhandlers.UserHandler,
	orgHandler *adminhandlers.OrgHandler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler { return middleware.RequestID()(next) })
	r.Use(func(next http.Handler) http.Handler { return middleware.RequestLogger(logger)(next) })
	r.Use(func(next http.Handler) http.Handler { return middleware.CORS(cfg.CORSAllowedOrigins)(next) })

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("server is awake"))
	})

	r.Route("/api", func(api chi.Router) {
		api.Use(func(next http.Handler) http.Handler {
			return middleware.RequireJSONContentType()(next)
		})

		// Public auth routes
		api.Post("/auth/signup", authHandler.Signup)
		api.Post("/auth/login", authHandler.Login)
		api.Post("/auth/refresh", authHandler.Refresh)
		api.Post("/auth/logout", authHandler.Logout)

		// Authenticated routes
		api.Group(func(authenticated chi.Router) {
			authenticated.Use(authMiddleware.Authenticate)

			// User routes
			authenticated.Get("/users/me", userHandler.GetMe)
			authenticated.Put("/users/me", userHandler.UpdateMe)

			// Organization routes
			authenticated.Post("/organizations", orgHandler.Create)
			authenticated.Get("/organizations", orgHandler.List)

			// Org-scoped routes (require membership)
			authenticated.Route("/organizations/{orgID}", func(orgRouter chi.Router) {
				orgRouter.Use(roleMW.RequireOrgMember)

				orgRouter.Get("/", orgHandler.Get)
				orgRouter.Get("/members", orgHandler.ListMembers)

				// Admin-only org actions
				orgRouter.Group(func(adminRouter chi.Router) {
					adminRouter.Use(roleMW.RequireOrgAdmin)

					adminRouter.Put("/", orgHandler.Update)
					adminRouter.Delete("/members/{userID}", orgHandler.RemoveMember)
				})
			})
		})
	})

	return r
}
