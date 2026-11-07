package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS configura las cabeceras de Cross-Origin Resource Sharing según los orígenes permitidos.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	originsMap := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originsMap[o] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if originsMap["*"] || originsMap[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Request-ID")
			c.Header("Access-Control-Expose-Headers", "Link, X-Request-ID")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
