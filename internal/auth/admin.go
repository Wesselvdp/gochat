package auth

import (
	"crypto/subtle"
	"fmt"
	"net"
	"os"

	"github.com/gin-gonic/gin"
)

// AdminAuthConfig holds configuration for admin authentication
type AdminAuthConfig struct {
	// APIKey is the secret key required for admin access
	APIKey string
	// AllowedIPs is a list of IP addresses permitted to access admin routes
	AllowedIPs []string
	// StrictMode enforces both API key and IP restrictions
	StrictMode bool
}

// AdminAuthMiddleware creates a middleware for securing admin routes
func AdminAuthMiddleware(config AdminAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Print("one")

		// Check API Key
		providedKey := c.GetHeader("X-Admin-Key")
		expectedKey := config.APIKey

		// Use constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(providedKey), []byte(expectedKey)) != 1 {
			c.JSON(403, gin.H{
				"error": "Invalid or missing admin key",
			})
			c.Abort()
			return
		}

		// IP Restriction (if enabled)
		if config.StrictMode || len(config.AllowedIPs) > 0 {
			clientIP := getClientIP(c)

			if !isIPAllowed(clientIP, config.AllowedIPs) {
				c.JSON(403, gin.H{
					"error": "Access from this IP is not permitted",
				})
				c.Abort()
				return
			}
		}
		fmt.Print("success")
		c.Next()
	}
}

// getClientIP retrieves the client's IP address
func getClientIP(c *gin.Context) string {
	// Try to get real IP from X-Forwarded-For or X-Real-IP headers
	ip := c.GetHeader("X-Forwarded-For")
	if ip == "" {
		ip = c.GetHeader("X-Real-IP")
	}

	// Fallback to remote address
	if ip == "" {
		ip = c.ClientIP()
	}

	return ip
}

// isIPAllowed checks if the client IP is in the allowed list
func isIPAllowed(clientIP string, allowedIPs []string) bool {
	// If no IPs are specified, allow all
	if len(allowedIPs) == 0 {
		return true
	}

	// Parse client IP
	parsedClientIP := net.ParseIP(clientIP)
	if parsedClientIP == nil {
		return false
	}

	// Check against allowed IPs
	for _, allowedIP := range allowedIPs {
		// Support both exact IPs and CIDR ranges
		if _, allowedIPNet, err := net.ParseCIDR(allowedIP); err == nil {
			// CIDR check
			if allowedIPNet.Contains(parsedClientIP) {
				return true
			}
		} else {
			// Exact IP check
			if allowedIP == clientIP {
				return true
			}
		}
	}

	return false
}

// LoadAdminAuthConfig loads admin authentication configuration
func LoadAdminAuthConfig() AdminAuthConfig {
	return AdminAuthConfig{
		APIKey: os.Getenv("ADMIN_API_KEY"),
		AllowedIPs: []string{
			"92.68.199.197",      // Breuer Intraval
			"178.198.205.242/32", // localhost IPv4
			"::1/128",            // localhost IPv6
		},
		StrictMode: true,
	}
}
