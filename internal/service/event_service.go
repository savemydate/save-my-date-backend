package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jomir/save-my-date-backend/internal/domain"
)

// Errores de negocio estandar para eventos
var (
	ErrEventNotFound = errors.New("evento no encontrado")
	ErrInvalidID     = errors.New("el identificador del evento no es valido")
)

// EventService orquesta la lógica de negocio para los eventos
type EventService struct {
	repo domain.EventRepository
}

// NewEventService crea una nueva instancia del servicio recibiendo la interfaz del repositorio
func NewEventService(repo domain.EventRepository) *EventService {
	return &EventService{
		repo: repo,
	}
}

// GetAll obtiene todos los eventos registrados
func (s *EventService) GetAll(ctx context.Context) ([]domain.Event, error) {
	events, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error en servicio al listar eventos: %w", err)
	}

	return events, nil
}

// GetByID busca un evento validando el identificador y retornando ErrEventNotFound si no existe
func (s *EventService) GetByID(ctx context.Context, id string) (*domain.Event, error) {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return nil, ErrInvalidID
	}

	event, err := s.repo.FindByID(ctx, cleanID)
	if err != nil {
		return nil, fmt.Errorf("Error en servicio al buscar evento: %w", err)
	}

	if event == nil {
		return nil, ErrEventNotFound
	}

	return event, nil
}

// GetActiveEvent obtiene el evento principal activo (Útil para a landing pública de la invitación)
func (s *EventService) GetActiveEvent(ctx context.Context) (*domain.Event, error) {
	event, err := s.repo.FindFirst(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error en el servicio al obtener el evento activo: %w", err)
	}

	if event == nil {
		return nil, ErrEventNotFound
	}

	return event, nil
}
