package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"time"

	"github.com/eridia/initium/backend/internal/adapter/handler"
	"github.com/eridia/initium/backend/internal/adapter/middleware"
	"github.com/eridia/initium/backend/internal/adapter/persistence"
	"github.com/eridia/initium/backend/internal/infra"
	"github.com/eridia/initium/backend/internal/infra/config"
	"github.com/eridia/initium/backend/internal/infra/database"
	"github.com/eridia/initium/backend/internal/infra/email"
	"github.com/eridia/initium/backend/internal/infra/google"
	"github.com/eridia/initium/backend/internal/infra/token"
	"github.com/eridia/initium/backend/internal/service"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	// Database
	db, err := database.NewPostgresDB(cfg.DatabaseDSN())
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}

	if err := database.RunMigrations(cfg.DatabaseURL(), "file://migrations"); err != nil {
		log.Fatalf("running migrations: %v", err)
	}

	// Infrastructure
	tokenGen, err := token.NewJWTGenerator(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath)
	if err != nil {
		log.Fatalf("initializing JWT generator: %v", err)
	}

	oauthVerifier := google.NewOAuthVerifier(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)

	emailSender, err := email.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.AppURL)
	if err != nil {
		log.Fatalf("initializing email sender: %v", err)
	}

	// Repositories
	userRepo := persistence.NewGormUserRepo(db)
	sessionRepo := persistence.NewGormSessionRepo(db)

	// Services
	authService := service.NewAuthService(userRepo, sessionRepo, oauthVerifier, emailSender, tokenGen)
	userService := service.NewUserService(userRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService, oauthVerifier, cfg.AppURL)
	mobileAuthHandler := handler.NewMobileAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

	// Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.AppURL},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/healthz", handler.Healthz)

	// Public routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/landing", handler.Landing)

		// Auth routes (rate limited)
		r.Route("/auth", func(r chi.Router) {
			r.Use(httprate.LimitByIP(10, time.Minute))

			r.Get("/google", authHandler.GoogleRedirect)
			r.Get("/google/callback", authHandler.GoogleCallback)
			r.Post("/magic-link", authHandler.RequestMagicLink)
			r.Get("/verify", authHandler.VerifyMagicLink)
			r.Post("/refresh", authHandler.RefreshTokens)
			r.Post("/mobile/google", mobileAuthHandler.GoogleIDToken)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(tokenGen, cfg.DevBypassAuth))

			r.Get("/me", userHandler.GetProfile)
			r.Patch("/me", userHandler.UpdateProfile)
			r.Post("/auth/logout", authHandler.Logout)
		})
	})

	slog.Info("configuration loaded",
		"env", cfg.AppEnv,
		"dev_bypass_auth", cfg.DevBypassAuth,
		"port", cfg.HTTPPort,
	)

	if err := infra.ServeHTTP(r, cfg.HTTPPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
