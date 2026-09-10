package middlewares

import (
	"crypto/rsa"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nvnamsss/unlimit/errors"
	"github.com/nvnamsss/unlimit/utility"
)

var (
	UserIDKey        = "user_id"
	RoleKey          = "role"
	UUIDKey          = "uuid"
	XUserIDHeader    = "X-User-Id"
	XUserRefIDHeader = "X-User-Ref-Id"
	XUserRoleHeader  = "X-User-Role"
)

// AuthConfig contains configuration for the auth middleware
//
// Using one of JWTSecret or PublicKey is required depending on the signing method used.
type AuthConfig struct {
	// JWTSecret is the secret key used for HMAC signing
	JWTSecret string
	// PublicKey is the RSA public key used for RSA signing
	PublicKey    *rsa.PublicKey
	PublicKeyFnc func() (*rsa.PublicKey, error)
	// PublicKeyByKIDFnc resolves an RSA public key by JWT kid.
	// This is intended for JWKS-style key caches that refresh out-of-band.
	PublicKeyByKIDFnc func(kid string) (*rsa.PublicKey, error)
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware creates a Gin middleware for JWT authentication
func AuthMiddleware(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ErrMissingAuthHeader)
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ErrInvalidAuthFormat)
			return
		}

		// Parse and validate the token
		tokenString := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			switch token.Method.(type) {
			case *jwt.SigningMethodHMAC:
				if config.JWTSecret == "" {
					return nil, errors.ErrInvalidToken
				}
				return []byte(config.JWTSecret), nil
			case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS:
				return resolveRSAPublicKey(token, config)
			default:
				return nil, errors.ErrInvalidToken
			}
		})

		if err != nil {
			if err == jwt.ErrTokenExpired {
				c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ErrTokenExpired)
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ErrInvalidToken)
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ErrInvalidToken)
			return
		}

		userIDStr := c.GetHeader(XUserIDHeader)
		userID, _ := strconv.ParseInt(userIDStr, 10, 64)
		// Add claims to context

		utility.SetUserRefIDToGinContext(c, claims.UserID)
		utility.SetUserIDToGinContext(c, userID)
		utility.SetRoleToGinContext(c, claims.Role)

		c.Next()
	}
}

func resolveRSAPublicKey(token *jwt.Token, config AuthConfig) (*rsa.PublicKey, error) {
	if config.PublicKeyByKIDFnc != nil {
		kid, _ := token.Header["kid"].(string)
		if kid != "" {
			return config.PublicKeyByKIDFnc(kid)
		}
	}

	if config.PublicKey != nil {
		return config.PublicKey, nil
	}

	if config.PublicKeyFnc != nil {
		return config.PublicKeyFnc()
	}

	return nil, errors.ErrInvalidToken
}

// AuthDebugMiddleware creates a Gin middleware for debug authentication bypassing JWT validation.
// It sets predefined authentication values in the context for development and testing purposes.
// Example usage:
//
//	if os.Getenv("ENV") == "development" {
//		router.Use(AuthDebugMiddleware())
//	} else {
//		router.Use(AuthMiddleware(authConfig))
//	}
//
// This middleware bypasses all JWT validation and sets a fixed admin role with debug UUID.
// It still respects the X-User-Id header for user identification.
// WARNING: This should NEVER be used in production environments as it provides unrestricted access.
func AuthDebugMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader(XUserIDHeader)
		userID, _ := strconv.ParseInt(userIDStr, 10, 64)
		userRefID := c.GetHeader(XUserRefIDHeader)
		userRole := c.GetHeader(XUserRoleHeader)

		utility.SetUserRefIDToGinContext(c, userRefID)
		utility.SetUserIDToGinContext(c, userID)
		utility.SetRoleToGinContext(c, userRole)
		c.Next()
	}
}
