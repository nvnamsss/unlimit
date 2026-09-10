package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nvnamsss/unlimit/errors"
)

// RoleConfig defines required roles for an endpoint
type RoleConfig struct {
	AllowedRoles []string // List of roles that can access the endpoint
}

// RequireRoles creates a middleware that checks if user has required roles
//
// Example usage:
//
//	router.GET("/admin",
//	    AuthMiddleware(authConfig),
//	    RequireRoles(RoleConfig{
//	        AllowedRoles: []string{RoleAdmin},
//	    }),
//	    adminHandler,
//	)
func RequireRoles(config RoleConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context (set by AuthMiddleware)
		role, exists := c.Get(RoleKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ErrUnauthorized)
			return
		}

		// Check if user's role is in allowed roles
		userRole := role.(string)
		for _, allowedRole := range config.AllowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		// User's role not allowed
		c.AbortWithStatusJSON(http.StatusForbidden, errors.ErrInsufficientPermissions)
	}
}

// Common role constants
const (
	RoleAdmin    = "admin"
	RoleUser     = "user"
	RoleOperator = "operator"
)
