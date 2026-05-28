package router

import (
	"time"

	"github.com/fastax/fastax-server/internal/shared/config"
	"github.com/fastax/fastax-server/internal/shared/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// Global middleware
	r.Use(middleware.CORS())
	r.Use(middleware.DetectLanguage())

	// Rate limiters
	ipLimiter := middleware.NewRateLimiter(cfg.RateLimit.IP, time.Minute)
	authLimiter := middleware.NewRateLimiter(cfg.RateLimit.Auth, time.Minute)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api")
	{
		// Auth endpoints (no JWT required)
		auth := api.Group("/auth")
		auth.Use(middleware.RateLimitAuth(authLimiter))
		{
			auth.POST("/register", placeholder)
			auth.POST("/login", placeholder)
			auth.POST("/refresh", placeholder)
			auth.POST("/send-code", placeholder)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthRequired(cfg.JWT.Secret))
		{
			tokens := protected.Group("/tokens")
			{
				tokens.GET("/products", placeholder)
				tokens.GET("/my", placeholder)
				tokens.POST("/buy", placeholder)
			}

			orders := protected.Group("/orders")
			{
				orders.GET("", placeholder)
			}
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(cfg.JWT.Secret), middleware.AdminRequired())
		{
			admin.GET("/dashboard/summary", placeholder)
			admin.GET("/users", placeholder)
			admin.GET("/suppliers", placeholder)
			admin.GET("/orders", placeholder)
		}
	}

	// OpenAI-compatible relay routes (S3+)
	v1 := r.Group("/v1")
	v1.Use(middleware.RateLimitIP(ipLimiter))
	{
		v1.POST("/chat/completions", placeholder)
		v1.POST("/embeddings", placeholder)
	}
}

func placeholder(c *gin.Context) {
	c.JSON(200, gin.H{"message": "not implemented yet"})
}
