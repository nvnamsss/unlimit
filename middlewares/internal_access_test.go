package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllowInternalCIDR(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		wantStatus int
		wantError  string
	}{
		{
			name:       "allows 10.0.0.0/8",
			remoteAddr: "10.1.2.3:1234",
			wantStatus: http.StatusOK,
		},
		{
			name:       "allows 172.16.0.0/12",
			remoteAddr: "172.16.1.1:443",
			wantStatus: http.StatusOK,
		},
		{
			name:       "allows 192.168.0.0/16",
			remoteAddr: "192.168.100.200:80",
			wantStatus: http.StatusOK,
		},
		{
			name:       "forbids public ip",
			remoteAddr: "8.8.8.8:53",
			wantStatus: http.StatusForbidden,
			wantError:  "internal access only",
		},
		{
			name:       "forbids invalid remote addr",
			remoteAddr: "not-an-ip",
			wantStatus: http.StatusForbidden,
			wantError:  "internal access only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			m := AllowInternalCIDR([]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			c.Request = req

			m(c)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus != http.StatusOK {
				var body map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &body)
				require.NoError(t, err)
				assert.Equal(t, tt.wantError, body["error"])
			}
		})
	}
}

func TestRequireClusterCaller(t *testing.T) {
	tests := []struct {
		name       string
		headerVal  string
		wantStatus int
		wantError  string
	}{
		{
			name:       "forbids missing header",
			headerVal:  "",
			wantStatus: http.StatusForbidden,
			wantError:  "private service access only",
		},
		{
			name:       "allows when header present",
			headerVal:  "scheduler",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			m := RequireClusterCaller()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.headerVal != "" {
				req.Header.Set("X-Caller-Service", tt.headerVal)
			}
			c.Request = req

			m(c)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus != http.StatusOK {
				var body map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &body)
				require.NoError(t, err)
				assert.Equal(t, tt.wantError, body["error"])
			}
		})
	}
}
