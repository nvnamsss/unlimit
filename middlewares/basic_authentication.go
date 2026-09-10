package middlewares

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nvnamsss/unlimit/errors"
)

var (
	BasicAuthUserKey = "basic_auth_user"
)

// BasicAuthConfig contains configuration for the basic auth middleware
type BasicAuthConfig struct {
	Username string
	Password string
}

// BasicAuthMiddleware creates a Gin middleware for HTTP Basic Authentication
func BasicAuthMiddleware(config BasicAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ErrMissingAuthHeader)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Basic" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ErrInvalidAuthFormat)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ErrInvalidAuthFormat)
			return
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ErrInvalidAuthFormat)
			return
		}

		username, password := pair[0], pair[1]
		if username != config.Username || password != config.Password {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ErrInvalidToken)
			return
		}

		c.Set(BasicAuthUserKey, username)
		c.Next()
	}
}
