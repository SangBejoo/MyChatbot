package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

type Middleware struct {
	jwtSecret   []byte
	rateLimiters map[string]*rate.Limiter
	mu          sync.Mutex
}

func NewMiddleware(secret string) *Middleware {
	return &Middleware{
		jwtSecret:    []byte(secret),
		rateLimiters: make(map[string]*rate.Limiter),
	}
}

func (m *Middleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
			c.Set("role", claims["role"])
		}

		c.Next()
	}
}

// RateLimitPerUser limits requests based on "user_id" from context (must follow AuthRequired)
func (m *Middleware) RateLimitPerUser(r rate.Limit, b int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			// Should not happen if AuthRequired is used, but safe fallback
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User identity not found for rate limiting"})
			return
		}

		key := strconv.FormatFloat(userID.(float64), 'f', 0, 64) // JWT numbers are float64 by default

		m.mu.Lock()
		limiter, exists := m.rateLimiters[key]
		if !exists {
			limiter = rate.NewLimiter(r, b)
			m.rateLimiters[key] = limiter
		}
		m.mu.Unlock()

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}

		c.Next()
	}
}

// CORSMiddleware allows Cross-Origin requests
func (m *Middleware) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// SecurityHeaders adds security headers to prevent common attacks
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		// Prevent clickjacking
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		// XSS Protection (legacy but still useful)
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		// Referrer policy
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Content Security Policy (basic)
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		
		c.Next()
	}
}

// RequestSizeLimiter limits request body size to prevent DoS
func RequestSizeLimiter(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
