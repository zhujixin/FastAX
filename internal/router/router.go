package router

import (
	"time"

	"github.com/fastax/fastax-server/internal/domain/token"
	"github.com/fastax/fastax-server/internal/domain/user"
	"github.com/fastax/fastax-server/internal/shared/cache"
	"github.com/fastax/fastax-server/internal/shared/config"
	"github.com/fastax/fastax-server/internal/shared/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handlers struct {
	User  *user.Handler
	Token *token.Handler
}

func NewHandlers(db *gorm.DB, redis *cache.RedisClient, cfg *config.Config) *Handlers {
	userSvc := user.NewService(db, redis, &cfg.JWT)
	tokenSvc := token.NewService(db)
	return &Handlers{
		User:  user.NewHandler(userSvc),
		Token: token.NewHandler(tokenSvc),
	}
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB, redis *cache.RedisClient, cfg *config.Config) {
	h := NewHandlers(db, redis, cfg)

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
			auth.POST("/register", h.User.Register)
			auth.POST("/login", h.User.Login)
			auth.POST("/refresh", h.User.RefreshToken)
			auth.POST("/send-code", h.User.SendCode)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthRequired(cfg.JWT.Secret))
		{
			protected.GET("/user/me", h.User.GetUser)

			tokens := protected.Group("/tokens")
			{
				tokens.GET("/products", h.Token.GetProducts)
				tokens.GET("/my", h.Token.GetMyTokens)
				tokens.POST("/buy", h.Token.Buy)
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
