package router

import (
	"net/http"
	"time"

	"github.com/fastax/fastax-server/internal/domain/byok"
	"github.com/fastax/fastax-server/internal/domain/commission"
	"github.com/fastax/fastax-server/internal/domain/cost"
	"github.com/fastax/fastax-server/internal/domain/enterprise"
	"github.com/fastax/fastax-server/internal/domain/guardrail"
	logpkg "github.com/fastax/fastax-server/internal/domain/log"
	"github.com/fastax/fastax-server/internal/domain/market"
	"github.com/fastax/fastax-server/internal/domain/notify"
	"github.com/fastax/fastax-server/internal/domain/order"
	"github.com/fastax/fastax-server/internal/domain/payment"
	"github.com/fastax/fastax-server/internal/domain/proxy"
	"github.com/fastax/fastax-server/internal/domain/risk"
	"github.com/fastax/fastax-server/internal/domain/stats"
	"github.com/fastax/fastax-server/internal/domain/token"
	"github.com/fastax/fastax-server/internal/domain/user"
	"github.com/fastax/fastax-server/internal/domain/vendor"
	"github.com/fastax/fastax-server/internal/shared/cache"
	"github.com/fastax/fastax-server/internal/shared/config"
	"github.com/fastax/fastax-server/internal/shared/i18n"
	"github.com/fastax/fastax-server/internal/shared/middleware"
	"github.com/fastax/fastax-server/internal/shared/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handlers struct {
	User      *user.Handler
	Token     *token.Handler
	Order     *order.Handler
	Payment   *payment.Handler
	Proxy     *proxy.Handler
	Vendor    *vendor.Handler
	Notify    *notify.Handler
	Guardrail *guardrail.Handler
	Risk      *risk.Handler
	BYOK       *byok.Handler
	Log        *logpkg.Handler
	Commission *commission.Handler
	Stats      *stats.Handler
	Enterprise *enterprise.Handler
	Cost       *cost.Handler
	Market     *market.Handler
	I18n       *i18n.Handler
}

func NewHandlers(db *gorm.DB, redis *cache.RedisClient, cfg *config.Config) *Handlers {
	userSvc := user.NewService(db, redis, &cfg.JWT)
	tokenSvc := token.NewService(db)
	orderSvc := order.NewService(db)
	paymentSvc := payment.NewService(db)
	proxySvc := proxy.NewService(db)
	vendorSvc := vendor.NewService(db)
	notifySvc := notify.NewService(db)
	guardrailSvc := guardrail.NewService(db, cfg.Guardrail.Mode)
	riskSvc := risk.NewService(db)
	byokSvc := byok.NewService(db)
	logSvc := logpkg.NewService(db)
	commissionSvc := commission.NewService(db)
	statsSvc := stats.NewService(db)
	enterpriseSvc := enterprise.NewService(db)
	costSvc := cost.NewService(db)
	marketSvc := market.NewService(db)
	i18nSvc := i18n.NewService(db)
	return &Handlers{
		User:      user.NewHandler(userSvc),
		Token:     token.NewHandler(tokenSvc),
		Order:     order.NewHandler(orderSvc, paymentSvc),
		Payment:   payment.NewHandler(paymentSvc),
		Proxy:     proxy.NewHandler(proxySvc),
		Vendor:    vendor.NewHandler(vendorSvc),
		Notify:    notify.NewHandler(notifySvc),
		Guardrail: guardrail.NewHandler(guardrailSvc),
		Risk:      risk.NewHandler(riskSvc),
		BYOK:      byok.NewHandler(byokSvc),
		Log:        logpkg.NewHandler(logSvc),
		Commission: commission.NewHandler(commissionSvc),
		Stats:      stats.NewHandler(statsSvc),
		Enterprise: enterprise.NewHandler(enterpriseSvc),
		Cost:       cost.NewHandler(costSvc),
		Market:     market.NewHandler(marketSvc),
		I18n:       i18n.NewHandler(i18nSvc),
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
		// Public routes
		api.GET("/tokens/products", h.Token.GetProducts)
		api.GET("/tokens/products/:id", h.Token.GetProduct)
		api.GET("/models", h.Market.ListModels)
		api.GET("/providers/health", h.Market.ListProviders)
		api.GET("/providers/:id/health", h.Market.GetProviderHealth)

		// i18n public routes
		i18nGroup := api.Group("/i18n")
		{
			i18nGroup.GET("/languages", h.I18n.ListLanguages)
			i18nGroup.GET("/translations/:locale", h.I18n.GetTranslations)
		}

		// Auth endpoints (no JWT required)
		auth := api.Group("/auth")
		auth.Use(middleware.RateLimitAuth(authLimiter))
		{
			auth.POST("/register", h.User.Register)
			auth.POST("/login", h.User.Login)
			auth.POST("/refresh", h.User.RefreshToken)
			auth.POST("/send-code", h.User.SendCode)
			auth.POST("/reset-password", h.User.ResetPassword)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthRequired(cfg.JWT.Secret, redis))
		{
			protected.POST("/auth/logout", h.User.Logout)
			protected.GET("/user/me", h.User.GetUser)
			protected.PUT("/user/language", h.User.UpdateLanguage)
			protected.GET("/tokens/my", h.Token.GetMyTokens)
			protected.GET("/tokens/my/usage", h.Token.GetUsageHistory)
			protected.POST("/tokens/buy", h.Token.Buy)
			protected.POST("/tokens/transfer", h.Token.Transfer)
			protected.POST("/tokens/extract", h.Token.Extract)
			protected.POST("/models/compare", h.Market.CompareModels)

			orders := protected.Group("/orders")
			{
				orders.POST("", h.Order.Create)
				orders.GET("", h.Order.List)
				orders.GET("/:id", h.Order.Get)
				orders.POST("/:id/cancel", h.Order.Cancel)
				orders.POST("/:id/refund", h.Order.RequestRefund)
			}

			payments := protected.Group("/payments")
			{
				payments.POST("", h.Payment.Create)
				payments.POST("/callback", h.Payment.Callback)
				payments.GET("/:order_id", h.Payment.GetPayment)
				payments.POST("/refunds", h.Payment.CreateRefund)
				payments.GET("/refunds", h.Payment.ListRefunds)
			}

			notifs := protected.Group("/notifications")
			{
				notifs.GET("", h.Notify.List)
				notifs.GET("/unread-count", h.Notify.UnreadCount)
				notifs.PUT("/:id/read", h.Notify.MarkRead)
				notifs.PUT("/read-all", h.Notify.MarkAllRead)
			}

			byokGroup := protected.Group("/byok")
			{
				byokGroup.GET("/keys", h.BYOK.ListKeys)
				byokGroup.POST("/keys", h.BYOK.AddKey)
				byokGroup.DELETE("/keys/:id", h.BYOK.DeleteKey)
				byokGroup.PUT("/keys/:id/status", h.BYOK.SetKeyStatus)
			}

			commissions := protected.Group("/commissions")
			{
				commissions.GET("", h.Commission.ListCommissions)
				commissions.GET("/total", h.Commission.GetTotal)
				commissions.POST("/withdraw", h.Commission.Withdraw)
			}

			statsGroup := protected.Group("/stats")
			{
				statsGroup.GET("/usage", h.Stats.GetUsage)
				statsGroup.GET("/consumption", h.Stats.GetConsumption)
				statsGroup.GET("/bills", h.Stats.GetBills)
				statsGroup.GET("/summary", h.Stats.GetSummary)
			}

			enterpriseGroup := protected.Group("/enterprise")
			{
				enterpriseGroup.POST("/sub-accounts", h.Enterprise.CreateSubAccount)
				enterpriseGroup.GET("/sub-accounts", h.Enterprise.ListSubAccounts)
				enterpriseGroup.PUT("/sub-accounts/:id/status", h.Enterprise.SetSubAccountStatus)
				enterpriseGroup.PUT("/sub-accounts/:id/quota", h.Enterprise.UpdateQuota)
				enterpriseGroup.GET("/usage", h.Enterprise.GetUsage)
				enterpriseGroup.GET("/sub-accounts/:id/usage", h.Enterprise.GetSubAccountUsage)
			}

			costGroup := protected.Group("/user")
			{
				costGroup.GET("/budget", h.Cost.GetBudget)
				costGroup.PUT("/budget", h.Cost.SetBudget)
				costGroup.GET("/cost-alerts", h.Cost.GetAlerts)
				costGroup.PUT("/cost-alerts", h.Cost.SetAlert)
			}
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(cfg.JWT.Secret, redis), middleware.AdminRequired())
		{
			admin.GET("/dashboard/summary", h.Stats.GetDashboardSummary)
			admin.GET("/users", h.User.ListUsers)
			admin.GET("/users/:id", h.User.GetUserDetail)
			admin.PUT("/users/:id/status", h.User.SetUserStatus)
			admin.PUT("/users/:id/level", h.User.SetUserLevel)
			admin.GET("/orders", h.Order.ListAdmin)
			admin.POST("/orders/:id/refund", h.Order.AdminRefund)

			// Product management
			admin.POST("/products", h.Token.CreateProduct)
			admin.PUT("/products/:id", h.Token.UpdateProduct)

			// Reports
			admin.GET("/reports/daily", h.Stats.GetDailyReport)
			admin.GET("/reports/monthly", h.Stats.GetMonthlyReport)

			// Supplier management
			admin.POST("/suppliers", h.Vendor.CreateSupplier)
			admin.GET("/suppliers", h.Vendor.ListSuppliers)
			admin.GET("/suppliers/:id", h.Vendor.GetSupplier)
			admin.PUT("/suppliers/:id", h.Vendor.UpdateSupplier)
			admin.PUT("/suppliers/:id/status", h.Vendor.SetSupplierStatus)

			// Vendor management
			admin.GET("/vendors", h.Vendor.ListVendors)
			admin.GET("/vendors/:id", h.Vendor.GetVendor)
			admin.POST("/vendors/:id/review", h.Vendor.ReviewVendor)
			admin.POST("/vendors/:id/suspend", h.Vendor.SuspendVendor)

			// Vendor product management
			admin.POST("/vendors/:vendor_id/products", h.Vendor.CreateProduct)
			admin.POST("/vendor-products/:id/review", h.Vendor.ReviewProduct)
			admin.GET("/vendors/:vendor_id/products", h.Vendor.ListProducts)

			// Risk management
			admin.GET("/risk/rules", h.Risk.ListRules)
			admin.POST("/risk/rules", h.Risk.CreateRule)
			admin.PUT("/risk/rules/:id/enabled", h.Risk.SetRuleEnabled)
			admin.GET("/risk/events", h.Risk.ListEvents)
			admin.PUT("/risk/events/:id/handle", h.Risk.HandleEvent)

			// Blacklist management
			admin.GET("/risk/blacklist", h.Risk.ListBlacklist)
			admin.POST("/risk/blacklist", h.Risk.AddBlacklist)
			admin.DELETE("/risk/blacklist/:id", h.Risk.RemoveBlacklist)

			// Commission management
			admin.POST("/commissions/:id/settle", h.Commission.Settle)

			// Guardrail management
			admin.GET("/guardrails/rules", h.Guardrail.ListRules)
			admin.POST("/guardrails/rules", h.Guardrail.CreateRule)
			admin.PUT("/guardrails/rules/:id/enabled", h.Guardrail.SetRuleEnabled)
			admin.GET("/guardrails/logs", h.Guardrail.ListLogs)
			admin.POST("/guardrails/detect", h.Guardrail.Detect)

			// Audit & call log management
			admin.GET("/audit/logs", h.Log.ListAuditLogs)
			admin.GET("/audit/export", h.Log.ExportAuditLogs)
			admin.GET("/call-logs", h.Log.ListCallLogs)

			// Notification template management
			admin.GET("/notifications/templates", h.Notify.ListTemplates)
			admin.POST("/notifications/templates", h.Notify.CreateTemplate)
			admin.PUT("/notifications/templates/:id", h.Notify.UpdateTemplate)

			// i18n management
			admin.GET("/i18n/languages", h.I18n.ListAllLanguages)
			admin.POST("/i18n/languages", h.I18n.CreateLanguage)
			admin.PUT("/i18n/languages/:locale", h.I18n.UpdateLanguage)
			admin.PUT("/i18n/default", h.I18n.SetDefaultLanguage)
		}

		// Vendor self-service (protected)
		vendorGroup := protected.Group("/vendor")
		{
			vendorGroup.POST("/apply", h.Vendor.Apply)
			vendorGroup.GET("/me", func(c *gin.Context) {
				userID, _ := c.Get("user_id")
				svc := vendor.NewService(db)
				resp, err := svc.GetVendorByUserID(userID.(uint))
				if err != nil {
					response.Error(c, http.StatusNotFound, response.CodeNotFound, err.Error())
					return
				}
				response.Success(c, resp)
			})
			vendorGroup.PUT("/profile", h.Vendor.UpdateProfile)
			vendorGroup.PUT("/products/:id", h.Vendor.UpdateProduct)
			vendorGroup.GET("/sales", h.Vendor.GetSales)
			vendorGroup.GET("/settlements", h.Vendor.GetSettlements)
			vendorGroup.POST("/settlements/:id/confirm", h.Vendor.ConfirmSettlement)
			vendorGroup.POST("/settlements/:id/withdraw", h.Vendor.RequestWithdrawal)
		}
	}

	// OpenAI-compatible relay routes
	v1 := r.Group("/v1")
	v1.Use(middleware.RateLimitIP(ipLimiter))
	{
		v1.POST("/chat/completions", h.Proxy.ChatCompletions)
		v1.POST("/messages", h.Proxy.ChatMessages) // Anthropic Messages API
		v1.POST("/images/generations", h.Proxy.ImageGenerations)
		v1.POST("/audio/speech", h.Proxy.AudioSpeech)
		v1.GET("/models", h.Proxy.ListModels)
	}
}

func placeholder(c *gin.Context) {
	c.JSON(200, gin.H{"message": "not implemented yet"})
}
