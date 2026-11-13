package health

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jomir/save-my-date-backend/pkg/response"
)

var startTime = time.Now()

// Pinger define la función mínima para validar la salud de una base de datos
type Pinger func(ctx context.Context) error

// StatusResponse representa la estructura de datos del estado operativo.
type StatusResponse struct {
	Status      string            `json:"status" example:"healthy"`
	Version     string            `json:"version" example:"1.0.0"`
	Environment string            `json:"environment" example:"development"`
	Timestamp   time.Time         `json:"timestamp" example:"2026-11-05T06:30:00Z"`
	Uptime      string            `json:"uptime" example:"1h 23m 45s"`
	Services    map[string]string `json:"services"`
}

// Handler gestiona los endpoints de health check.
type Handler struct {
	version     string
	environment string
	dbPinger    Pinger
}

// NewHandler crea una nueva instancia del Handler de salud.
func NewHandler(version, environment string, dbPinger Pinger) *Handler {
	return &Handler{
		version:     version,
		environment: environment,
		dbPinger:    dbPinger,
	}
}

// Check godoc
// @Summary      Comprobación del estado del servicio
// @Description  Retorna el estado operativo de la API, versión, tiempo activo y componentes auxiliares.
// @Tags         Health
// @Produce      json
// @Success      200  {object}  response.Response{data=health.StatusResponse}
// @Router       /health [get]
// @Router       /api/v1/health [get]
func (h *Handler) Check(c *gin.Context) {
	uptimeDuration := time.Since(startTime).Round(time.Second)

	dbStatus := "operational"
	if h.dbPinger != nil {
		if err := h.dbPinger(c.Request.Context()); err != nil {
			dbStatus = "degraded"
		}
	}

	status := StatusResponse{
		Status:      "healthy",
		Version:     h.version,
		Environment: h.environment,
		Timestamp:   time.Now().UTC(),
		Uptime:      formatDuration(uptimeDuration),
		Services: map[string]string{
			"api":      "operational",
			"database": dbStatus,
		},
	}

	response.OK(c, status)
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
}
