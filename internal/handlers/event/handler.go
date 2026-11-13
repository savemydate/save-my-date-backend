package event

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jomir/save-my-date-backend/internal/service"
	"github.com/jomir/save-my-date-backend/pkg/response"
)

// Handler gestiona los endpoints REST de eventos
type Handler struct {
	service *service.EventService
}

// NewHandler crea una nueva instancia del Handler de eventos
func NewHandler(s *service.EventService) *Handler {
	return &Handler{
		service: s,
	}
}

// GetAll godoc
// @Summary					Listar eventos
// @Description			Obtiene la lista de todos los eventos configurados en la plataforma
// @Tags						Events
// @Produce					json
// @Success					200  {object}  response.Response{data=[]domain.Event}
// @Failure					500  {object}  response.Response
// @Router					/api/v1/events [get]
func (h *Handler) GetAll(c *gin.Context) {
	events, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Error al consultar los eventos")
		return
	}

	response.OK(c, events)
}

// GetByID godoc
// @Summary    			Consultar evento por ID
// @Description			Retorna los detalles completos de un evento según su identificador único.
// @Tags       			Events
// @Produce    			json
// @Param      			id   path      string  true  "ID del Evento o Mongo ObjectID"
// @Success    			200  {object}  response.Response{data=domain.Event}
// @Failure    			400  {object}  response.Response
// @Failure    			404  {object}  response.Response
// @Failure    			500  {object}  response.Response
// @Router     			/api/v1/events/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	event, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrInvalidID) {
			response.Error(c, http.StatusBadRequest, "INVALID_ID", "El identificador proporcionado no es válido")
			return
		}

		if errors.Is(err, service.ErrEventNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "No se encontró ningún evento con este identificador")
			return
		}

		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al recuperar el evento")
		return
	}

	response.OK(c, event)
}

// GetActive godoc
// @Summary    			Obtener evento activo principal
// @Description			Retorna la configuración del evento principal para renderizar la invitación web pública.
// @Tags       			Events
// @Produce    			json
// @Success    			200  {object}  response.Response{data=domain.Event}
// @Failure    			404  {object}  response.Response
// @Failure    			500  {object}  response.Response
// @Router     			/api/v1/events/active [get]
func (h *Handler) GetActive(c *gin.Context) {
	event, err := h.service.GetActiveEvent(c.Request.Context())
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "No hay ningún evento activo configurado actualmente")
			return
		}

		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al recuperar el evento activo")
		return
	}

	response.OK(c, event)
}
