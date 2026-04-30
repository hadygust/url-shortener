package ratelimit

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hadygust/url-shortener/internal/dto"
)

// UserRateLimitMiddleware limits requests per user (e.g., URL creation)
// Limit: 10 requests per minute
func (rl *RateLimiter) UserRateLimitMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (should be set by auth middleware)
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
			c.Abort()
			return
		}

		userID := user.(dto.UserResponse).ID.String()
		key := "ratelimit:user:" + userID

		allowed, _, err := rl.AllowRequest(key, maxRequests, window)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limit check failed"})
			c.Abort()
			return
		}

		if !allowed {
			c.Header("X-RateLimit-Limit", "10")
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", time.Now().Add(window).Format(time.RFC3339))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "too many requests",
				"message": "URL creation limit exceeded. Maximum 10 requests per minute.",
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining, _ := rl.GetRemaining(key, maxRequests, window)
		c.Header("X-RateLimit-Limit", "10")
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", time.Now().Add(window).Format(time.RFC3339))

		c.Next()
	}
}

// IPRateLimitMiddleware limits requests per IP (e.g., redirects)
// Limit: 60 requests per minute
func (rl *RateLimiter) IPRateLimitMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := "ratelimit:ip:" + clientIP

		allowed, _, err := rl.AllowRequest(key, maxRequests, window)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limit check failed"})
			c.Abort()
			return
		}

		if !allowed {
			c.Header("X-RateLimit-Limit", "60")
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", time.Now().Add(window).Format(time.RFC3339))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "too many requests",
				"message": "Redirect limit exceeded. Maximum 60 requests per minute.",
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining, _ := rl.GetRemaining(key, maxRequests, window)
		c.Header("X-RateLimit-Limit", "60")
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", time.Now().Add(window).Format(time.RFC3339))

		c.Next()
	}
}
