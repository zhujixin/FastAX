package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/fastax/fastax-server/internal/router"
	"github.com/fastax/fastax-server/internal/shared/cache"
	"github.com/fastax/fastax-server/internal/shared/config"
	"github.com/fastax/fastax-server/internal/shared/model"
	"github.com/gin-gonic/gin"
)

func getConfigPath() string {
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		return p
	}
	return "config.yaml"
}

// validateConfig performs startup safety checks on critical configuration.
// It exits fatally if required values are missing or invalid in production mode.
func validateConfig(cfg *config.Config) {
	mode := cfg.Server.Mode
	isProd := mode == gin.ReleaseMode

	// JWT secrets MUST be configured in production
	if cfg.JWT.Secret == "" {
		if isProd {
			log.Fatal("FATAL: jwt.secret is empty — must be configured in config.yaml for production")
		}
		log.Println("WARNING: jwt.secret is empty — using insecure default (dev only)")
	}

	// Refresh token secret: warn if not configured independently
	if cfg.JWT.RefreshSecret == "" {
		if isProd {
			log.Println("WARNING: jwt.refresh_secret is empty — using derived key (secret+\"-refresh\")")
		}
	}

	// Encryption key: validate if configured
	if cfg.Security.EncryptionKey != "" {
		key := cfg.GetEncryptionKey()
		if key == nil {
			log.Fatal("FATAL: security.encryption_key is configured but cannot be decoded as 32-byte hex/base64")
		}
		if isProd {
			log.Println("AES-256 encryption key configured — API keys will be stored encrypted")
		}
	}
}

func main() {
	// Load config
	cfgPath := getConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate critical configuration
	validateConfig(cfg)

	// Initialize database
	db, err := model.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// Initialize Redis (optional, warn if unavailable)
	redisClient, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Printf("WARNING: Redis not available (running without cache): %v", err)
		redisClient = nil
	}
	_ = redisClient

	// Create Gin engine
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Register routes with service injection
	router.RegisterRoutes(r, db, redisClient, cfg)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("FastAX server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
