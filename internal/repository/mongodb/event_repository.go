package mongodb

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/jomir/save-my-date-backend/internal/domain"
)

const eventsCollectionName = "events"

// eventDocument es un struct auxiliar para leer de Mongo sin conflictos de tipo en _id
type eventDocument struct {
	domain.Event `bson:",inline"`
	RawID        any    `bson:"_id,omitempty"`
	CustomID     string `bson:"id,omitempty"`
}

// toDomain transforma el documento BSON de Mongo a la entidad limpia de Dominio.
func (doc *eventDocument) toDomain() domain.Event {
	evt := doc.Event

	// Si el documento ya traia un campo "id" en string, lo preservamos
	if doc.CustomID != "" {
		evt.ID = doc.CustomID

		return evt
	}

	// Si el _id de Mongo es un ObjectID nativo, lo convertimos a su representación hexadecimal en string
	switch id := doc.RawID.(type) {
	case bson.ObjectID:
		evt.ID = id.Hex()

	case string:
		evt.ID = id
	}

	return evt
}

// EventRepository implementa domain.EventRepository para MongoDB
type EventRepository struct {
	collection *mongo.Collection
}

// NewEventRepository es el constructor que inicializa el repositorio con la colección correspondiente
func NewEventRepository(db *mongo.Database) *EventRepository {
	return &EventRepository{
		collection: db.Collection(eventsCollectionName),
	}
}

// FindAll obtiene todos los eventos registrados en la colección
func (r *EventRepository) FindAll(ctx context.Context) ([]domain.Event, error) {
	// Filtro vacio: equivale a db.events.find({})
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("Error al consultar eventos en MongoDB: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []eventDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("Error al decodificar eventos: %w", err)
	}

	// Mapear documentos a entidades de dominio
	events := make([]domain.Event, len(docs))
	for i := range docs {
		events[i] = docs[i].toDomain()
	}

	return events, nil
}

// FindByID busca un evento por su ID (soporta tanto ObjectID de Mongo como ID en texto)
func (r *EventRepository) FindByID(ctx context.Context, id string) (*domain.Event, error) {
	var orFilters []bson.M

	// Si el ID es un ObjectID hexadecimal válido (24 caracteres), lo agregamos al filtro
	if objID, err := bson.ObjectIDFromHex(id); err == nil {
		orFilters = append(orFilters, bson.M{"_id": objID})
	}

	//También buscamos por _id como string o por un campo "id" secundario
	orFilters = append(orFilters, bson.M{"_id": id}, bson.M{"id": id})

	filter := bson.M{"$or": orFilters}

	var doc eventDocument
	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // No encontrado, retornamos sin error catastrófico
		}

		return nil, fmt.Errorf("Error al buscar evento por id (%s): %w", id, err)
	}

	event := doc.toDomain()
	return &event, nil
}

// FindFirst obtiene el primer evento registrado en la colección
func (r *EventRepository) FindFirst(ctx context.Context) (*domain.Event, error) {
	var doc eventDocument
	err := r.collection.FindOne(ctx, bson.M{}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, fmt.Errorf("Error al obtener el primer evento: %w", err)
	}

	event := doc.toDomain()
	return &event, nil
}
