package middlewares

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicAuthMiddleware(t *testing.T) {
	type args struct {
		authHeader string
	}

	validUsername := "testuser"
	validPassword := "testpass"
	validCredentials := base64.StdEncoding.EncodeToString([]byte(validUsername + ":" + validPassword))
	validAuthHeader := "Basic " + validCredentials

	tests := []struct {
		name       string
		args       args
		wantStatus int
		wantErr    bool
		validate   func(*testing.T, *gin.Context)
	}{
		{
			name: "should authorize with valid credentials",
			args: args{
				authHeader: validAuthHeader,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
			validate: func(t *testing.T, c *gin.Context) {
				user, exists := c.Get(BasicAuthUserKey)
				require.True(t, exists)
				assert.Equal(t, validUsername, user)
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
				authHeader: "Bearer sometoken",
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name: "should fail with invalid base64",
			args: args{
				authHeader: "Basic invalidbase64==",
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name: "should fail with invalid credentials format",
			args: args{
				authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("no_colon")),
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name: "should fail with wrong username",
			args: args{
				authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("wrong:"+validPassword)),
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name: "should fail with wrong password",
			args: args{
				authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte(validUsername+":wrongpass")),
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			config := BasicAuthConfig{
				Username: validUsername,
				Password: validPassword,
			}
			m := BasicAuthMiddleware(config)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.args.authHeader != "" {
				req.Header.Set("Authorization", tt.args.authHeader)
			}
			c.Request = req

			m(c)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.validate != nil && !tt.wantErr {
				tt.validate(t, c)
			}
		})
	}
}
