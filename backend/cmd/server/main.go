package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/eridia/initium/backend/internal/adapter/handler"
	"github.com/eridia/initium/backend/internal/adapter/persistence"
	"github.com/eridia/initium/backend/internal/app"
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

	r := app.NewRouter(app.RouterDeps{
		Auth:       authHandler,
		MobileAuth: mobileAuthHandler,
		User:       userHandler,
		TokenGen:   tokenGen,
		RoleLookup: roleLookup,
		DB:         db,
		AppEnv:     cfg.AppEnv,
		AppURL:     cfg.AppURL,
		DevBypass:  cfg.DevBypassAuth,
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
