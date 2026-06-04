// JWT 认证中间件
//
// 本文件实现了基于 JWT 的用户认证和授权：
//   - Claims: JWT 声明结构体，包含用户 ID、角色等信息
//   - AuthRequired: 认证中间件，验证 JWT Token 的有效性
//   - AdminRequired: 管理员权限中间件，检查用户是否为管理员
//   - GenerateToken: 生成 JWT Token（Access Token + Refresh Token）
//   - RefreshToken: 刷新 Access Token
package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fastax/fastax-server/internal/shared/cache"
	"github.com/fastax/fastax-server/internal/shared/constants"
	"github.com/fastax/fastax-server/internal/shared/model"
	"github.com/fastax/fastax-server/internal/shared/response"
	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 声明结构体，包含用户 ID、角色等信息
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthRequired validates JWT and sets user info in context.
// redis is optional (may be nil); when present it is used to check the logout blacklist.
func AuthRequired(secret string, redis *cache.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			response.Error(c, http.StatusUnauthorized, response.CodeTokenExpired, "missing authorization token")
			return
		}

		claims, err := parseJWT(tokenStr, secret)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, response.CodeTokenExpired, "invalid or expired token")
			return
		}

		// Check logout blacklist (skip if Redis unavailable)
		if redis != nil {
			blacklistKey := cache.BlacklistKey(claims.UserID)
			val, err := redis.Get(blacklistKey)
			if err == nil && val != "" {
				logoutAt, err := strconv.ParseInt(val, 10, 64)
				if err == nil && claims.IssuedAt != nil {
					// IssuedAt.Unix() is int64; logoutAt is Unix seconds
					if claims.IssuedAt.Unix() <= logoutAt {
						response.Error(c, http.StatusUnauthorized, response.CodeTokenExpired, "token has been invalidated by logout")
						return
					}
				}
			}
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// AdminRequired ensures the user has admin role
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != constants.RoleAdmin && role != constants.RoleSuperAdmin {
			response.Error(c, http.StatusForbidden, response.CodePermissionDeny)
			return
		}
		c.Next()
	}
}

// RoleRequired checks for specific role(s)
func RoleRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}
		response.Error(c, http.StatusForbidden, response.CodePermissionDeny)
	}
}

func extractToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return ""
}

// TokenAuthRequired validates API keys (product tokens) for proxy /v1 routes.
// Accepts "Bearer <token_id>" or "Bearer sk-<token_id>" format.
func TokenAuthRequired(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			response.Error(c, http.StatusUnauthorized, response.CodeTokenExpired, "missing api key")
			return
		}

		// Support both "<id>" and "sk-<id>" formats
		tokenStr = strings.TrimPrefix(tokenStr, "sk-")
		tokenID, err := strconv.ParseUint(tokenStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, response.CodeTokenExpired, "invalid api key format")
			return
		}

		var token model.UserToken
		if err := db.First(&token, tokenID).Error; err != nil {
			response.Error(c, http.StatusUnauthorized, response.CodeTokenExpired, "invalid api key")
			return
		}

		if token.Status != 1 {
			response.Error(c, http.StatusForbidden, response.CodeTokenExpired, "token disabled")
			return
		}

		if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
			response.Error(c, http.StatusForbidden, response.CodeTokenExpired, "token expired")
			return
		}

		c.Set("user_id", token.UserID)
		c.Set("token_id", token.ID)
		c.Next()
	}
}

func GenerateAccessToken(userID uint, role, secret string, expiry time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "fastax",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(userID uint, refreshSecret string, expiry time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "fastax-refresh",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(refreshSecret))
}

func parseJWT(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer to prevent cross-issuer token acceptance
	if claims.Issuer != "" && claims.Issuer != "fastax" && claims.Issuer != "fastax-refresh" {
		return nil, fmt.Errorf("unexpected issuer: %s", claims.Issuer)
	}

	return claims, nil
}
