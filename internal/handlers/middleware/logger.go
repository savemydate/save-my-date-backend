package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger middleware para logging estructurado con slog y trazabilidad mediante X-Request-ID.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			requestID = hex.EncodeToString(b)
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		logAttrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", statusCode),
			slog.Float64("duration_ms", float64(duration.Microseconds())/1000.0),
			slog.String("client_ip", c.ClientIP()),
		}

		if len(c.Errors) > 0 {
			logAttrs = append(logAttrs, slog.String("error", c.Errors.String()))
		}

		switch {
		case statusCode >= 500:
			slog.Error("HTTP Request Error", logAttrs...)
		case statusCode >= 400:
			slog.Warn("HTTP Request Warning", logAttrs...)
		default:
			slog.Info("HTTP Request Completed", logAttrs...)
		}
	}
}
