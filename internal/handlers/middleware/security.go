package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders inyecta cabeceras de seguridad HTTP defensivas recomendadas por OWASP.
func SecurityHeaders(isProduction bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Cabeceras universales
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-XSS-Protection", "1; mode=block")

		if isProduction {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Para rutas que no sean swagger, aplicar CSP estricto
		if !strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			c.Header("Content-Security-Policy", "default-src 'self'")
		}

		c.Next()
	}
}
