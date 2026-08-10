package middlewares

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AllowInternalCIDR only allows requests coming from RFC1918 private network ranges.
//
// It checks c.Request.RemoteAddr and allows:
// - 10.0.0.0/8
// - 172.16.0.0/12
// - 192.168.0.0/16
func AllowInternalCIDR(cidrs []string) gin.HandlerFunc {
	allowed := []*net.IPNet{}
	for _, cidr := range cidrs {
		_, n, err := net.ParseCIDR(cidr)
		if err == nil {
			allowed = append(allowed, n)
		}
	}

	return func(c *gin.Context) {
		host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			host = c.Request.RemoteAddr
		}
		ip := net.ParseIP(host)
		for _, n := range allowed {
			if ip != nil && n.Contains(ip) {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "internal access only"})
	}
}

// RequireClusterCaller only allows requests that provide X-Caller-Service header.
func RequireClusterCaller() gin.HandlerFunc {
	return func(c *gin.Context) {
		caller := strings.ToLower(c.GetHeader("X-Caller-Service"))
		if caller == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "private service access only"})
			return
		}
		c.Next()
	}
}
