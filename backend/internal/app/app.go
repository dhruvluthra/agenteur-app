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
	invitationRepo := adminservices.NewInvitationRepository()
	emailService := adminservices.NewConsoleEmailService()
	orgService := adminservices.NewOrgService(pool, orgRepo, membershipRepo)
	invitationService := adminservices.NewInvitationService(pool, invitationRepo, membershipRepo, emailService, userRepo, cfg.InviteBaseURL, cfg.InviteTokenTTL)
	orgHandler := adminhandlers.NewOrgHandler(orgService)
	invitationHandler := adminhandlers.NewInvitationHandler(invitationService, orgService)
	adminHandler := adminhandlers.NewAdminHandler(pool, userService)
	roleMW := adminhandlers.NewRoleMiddleware(pool, membershipRepo, userRepo)

	server := &http.Server{
		Addr:    cfg.Port,
		Handler: NewRouter(&RouterDeps{
			Config:            cfg,
			Logger:            logger,
			AuthMiddleware:    authMiddleware,
			RoleMiddleware:    roleMW,
			AuthHandler:       authHandler,
			UserHandler:       userHandler,
			OrgHandler:        orgHandler,
			InvitationHandler: invitationHandler,
			AdminHandler:      adminHandler,
		}),
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

type RouterDeps struct {
	Config            *config.Config
	Logger            *slog.Logger
	AuthMiddleware    *authhandlers.AuthMiddleware
	RoleMiddleware    *adminhandlers.RoleMiddleware
	AuthHandler       *authhandlers.AuthHandler
	UserHandler       *authhandlers.UserHandler
	OrgHandler        *adminhandlers.OrgHandler
	InvitationHandler *adminhandlers.InvitationHandler
	AdminHandler      *adminhandlers.AdminHandler
}

func NewRouter(deps *RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger(deps.Logger))
	r.Use(middleware.CORS(deps.Config.CORSAllowedOrigins))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("server is awake"))
	})

	r.Route("/api", func(api chi.Router) {
		api.Use(middleware.RequireJSONContentType())

		// Public auth routes
		api.Post("/auth/signup", deps.AuthHandler.Signup)
		api.Post("/auth/login", deps.AuthHandler.Login)
		api.Post("/auth/refresh", deps.AuthHandler.Refresh)
		api.Post("/auth/logout", deps.AuthHandler.Logout)

		// Public invitation view (token is the auth)
		api.Get("/invitations/{token}", deps.InvitationHandler.GetByToken)

		// Authenticated routes
		api.Group(func(authenticated chi.Router) {
			authenticated.Use(deps.AuthMiddleware.Authenticate)

			// User routes
			authenticated.Get("/users/me", deps.UserHandler.GetMe)
			authenticated.Put("/users/me", deps.UserHandler.UpdateMe)

			// Accept invitation (authenticated)
			authenticated.Post("/invitations/{token}/accept", deps.InvitationHandler.Accept)

			// Organization routes
			authenticated.Post("/organizations", deps.OrgHandler.Create)
			authenticated.Get("/organizations", deps.OrgHandler.List)

			// Superadmin routes
			authenticated.Route("/admin", func(adminRouter chi.Router) {
				adminRouter.Use(deps.RoleMiddleware.RequireSuperadmin)

				adminRouter.Get("/users", deps.AdminHandler.ListUsers)
				adminRouter.Put("/users/{userID}/superadmin", deps.AdminHandler.ToggleSuperadmin)
			})

			// Org-scoped routes (require membership)
			authenticated.Route("/organizations/{orgID}", func(orgRouter chi.Router) {
				orgRouter.Use(deps.RoleMiddleware.RequireOrgMember)

				orgRouter.Get("/", deps.OrgHandler.Get)
				orgRouter.Get("/members", deps.OrgHandler.ListMembers)

				// Admin-only org actions
				orgRouter.Group(func(adminRouter chi.Router) {
					adminRouter.Use(deps.RoleMiddleware.RequireOrgAdmin)

					adminRouter.Put("/", deps.OrgHandler.Update)
					adminRouter.Delete("/members/{userID}", deps.OrgHandler.RemoveMember)
					adminRouter.Post("/invitations", deps.InvitationHandler.Create)
				})
			})
		})
	})

	return r
}
