package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

func main() {
	// Load config
	cfgPath := getConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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
	_ = redisClient // available for injection into domain services

	// Create Gin engine
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Register routes (pass db and cfg for handler injection)
	router.RegisterRoutes(r, db, cfg)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
