package middlewares

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	// Test configuration
	type args struct {
		authHeader string
		userID     int64
	}

	// Sample test data
	jwtSecret := "test-secret"
	validUserID := int64(123)
	validRole := "user"

	// Create a valid JWT token for testing
	createValidToken := func() string {
		claims := Claims{
			UserID: "123",
			Role:   validRole,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Subject:   "test-subject",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte(jwtSecret))
		return "Bearer " + tokenString
	}

	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantErr    bool
		setup      func(*gin.Context)
		validate   func(*testing.T, *gin.Context)
	}{
		{
			name: "should authorize with valid token",
			args: args{
				authHeader: createValidToken(),
				userID:     validUserID,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
			setup: func(c *gin.Context) {
			},
			validate: func(t *testing.T, c *gin.Context) {
				// Verify context values were set correctly
				userID, exists := c.Get(UserIDKey)
				require.True(t, exists)
				assert.Equal(t, validUserID, userID)

				role, exists := c.Get(RoleKey)
				require.True(t, exists)
				assert.Equal(t, validRole, role)
			},
		},
		{
			name: "should fail with missing auth header",
			args: args{
				authHeader: "",
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name: "should fail with invalid auth format",
			args: args{
				authHeader: "InvalidFormat token",
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name: "should fail with invalid token",
			args: args{
				authHeader: "Bearer invalid.token.here",
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name: "should fail with expired token",
			args: args{
				authHeader: func() string {
					claims := Claims{
						UserID: "123",
						Role:   validRole,
						RegisteredClaims: jwt.RegisteredClaims{
							ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
						},
					}
					token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
					tokenString, _ := token.SignedString([]byte(jwtSecret))
					return "Bearer " + tokenString
				}(),
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Configure auth middleware
			authConfig := AuthConfig{
				JWTSecret: jwtSecret,
			}
			m := AuthMiddleware(authConfig)
			// Setup test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.args.authHeader != "" {
				req.Header.Set("Authorization", tt.args.authHeader)
			}
			req.Header.Set(XUserIDHeader, strconv.FormatInt(tt.args.userID, 10))

			// Setup any additional context
			if tt.setup != nil {
				tt.setup(c)
			}
			c.Request = req
			// Execute request
			// r.ServeHTTP(w, req)
			m(c)

			// Verify response
			assert.Equal(t, tt.wantStatus, w.Code)

			// Run additional validations
			if tt.validate != nil && !tt.wantErr {
				tt.validate(t, c)
			}
		})
	}
}

func TestAuthMiddleware_RSAWithKIDResolver(t *testing.T) {
	gin.SetMode(gin.TestMode)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	claims := Claims{
		UserID: "456",
		Role:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "test-subject-rsa",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "key-1"
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	m := AuthMiddleware(AuthConfig{
		PublicKeyByKIDFnc: func(kid string) (*rsa.PublicKey, error) {
			if kid == "key-1" {
				return &privateKey.PublicKey, nil
			}
			return nil, assert.AnError
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set(XUserIDHeader, "456")
	c.Request = req

	m(c)

	assert.Equal(t, http.StatusOK, w.Code)

	userID, exists := c.Get(UserIDKey)
	require.True(t, exists)
	assert.Equal(t, int64(456), userID)

	role, exists := c.Get(RoleKey)
	require.True(t, exists)
	assert.Equal(t, "admin", role)
}
