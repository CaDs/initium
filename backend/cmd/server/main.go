package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"

	"github.com/eridia/initium/backend/internal/adapter/handler"
	"github.com/eridia/initium/backend/internal/adapter/middleware"
	"github.com/eridia/initium/backend/internal/adapter/persistence"
	"github.com/eridia/initium/backend/internal/infra"
	"github.com/eridia/initium/backend/internal/infra/config"
	"github.com/eridia/initium/backend/internal/infra/cron"
	"github.com/eridia/initium/backend/internal/infra/database"
	"github.com/eridia/initium/backend/internal/infra/email"
	"github.com/eridia/initium/backend/internal/infra/google"
	"github.com/eridia/initium/backend/internal/infra/token"
	"github.com/eridia/initium/backend/internal/infra/worker"
	"github.com/eridia/initium/backend/internal/service"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	_ = godotenv.Load() // ignore error — .env is optional (e.g., in Docker)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("loading config", "error", err)
		os.Exit(1)
	}

	// Database
	db, err := database.NewPostgresDB(cfg.DatabaseDSN())
	if err != nil {
		slog.Error("connecting to database", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrations(cfg.DatabaseURL(), "file://migrations"); err != nil {
		slog.Error("running migrations", "error", err)
		os.Exit(1)
	}

	// Infrastructure
	tokenGen, err := token.NewJWTGenerator(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath)
	if err != nil {
		slog.Error("initializing JWT generator", "error", err)
		os.Exit(1)
	}

	oauthVerifier := google.NewOAuthVerifier(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)

	smtpSender, err := email.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom, cfg.AppURL, cfg.AppDeepScheme)
	if err != nil {
		slog.Error("initializing email sender", "error", err)
		os.Exit(1)
	}

	// Worker pool — async email delivery
	workerPool := worker.New()
	emailSender := email.NewAsyncSender(workerPool, smtpSender)

	// Repositories
	userRepo := persistence.NewGormUserRepo(db)
	sessionRepo := persistence.NewGormSessionRepo(db)

	// Services
	authService := service.NewAuthService(userRepo, sessionRepo, oauthVerifier, emailSender, tokenGen)
	userService := service.NewUserService(userRepo)

	// Cron scheduler — periodic cleanup tasks
	scheduler := cron.New()
	scheduler.Every(time.Hour, func(ctx context.Context) {
		n, err := sessionRepo.DeleteExpiredMagicLinks(ctx)
		if err != nil {
			slog.Error("magic link cleanup failed", "error", err)
			return
		}
		slog.Info("magic link cleanup complete", "deleted", n)
	})
	scheduler.Start(context.Background())

	// Handlers
	secureCookies := cfg.AppEnv != "development"
	authHandler := handler.NewAuthHandler(authService, oauthVerifier, cfg.AppURL, secureCookies)
	mobileAuthHandler := handler.NewMobileAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

	// Role lookup callback for RequireRole middleware (avoids importing repo into middleware).
	roleLookup := func(ctx context.Context, userID string) (string, error) {
		u, err := userRepo.FindByID(ctx, userID)
		if err != nil {
			return "", err
		}
		return u.Role, nil
	}

	// Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.AccessLog)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestSize(1 << 20)) // 1 MiB body limit
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.AppURL},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health / readiness checks
	r.Get("/healthz", handler.Healthz)
	r.Get("/readyz", handler.Readyz(db))

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
			r.Post("/mobile/verify", mobileAuthHandler.VerifyMagicLink)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(tokenGen, cfg.DevBypassAuth))

			r.Get("/me", userHandler.GetProfile)
			r.Patch("/me", userHandler.UpdateProfile)
			r.Post("/auth/logout", authHandler.Logout)
			r.Post("/auth/logout-all", authHandler.LogoutAll)
		})

		// Admin-only routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(tokenGen, cfg.DevBypassAuth))
			r.Use(middleware.RequireRole("admin", roleLookup))

			r.Get("/admin/ping", func(w http.ResponseWriter, req *http.Request) {
				handler.JSON(w, req, http.StatusOK, map[string]string{"role": "admin"})
			})
		})
	})

	slog.Info("configuration loaded",
		"env", cfg.AppEnv,
		"dev_bypass_auth", cfg.DevBypassAuth,
		"port", cfg.HTTPPort,
	)

	if err := infra.ServeHTTP(r, cfg.HTTPPort,
		func(_ context.Context) { scheduler.Stop() },
		func(_ context.Context) { workerPool.Close() },
	); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
