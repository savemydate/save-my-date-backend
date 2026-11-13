package domain

import (
	"context"
)

// LocationInfo representa la informacion geografica y de horario de una ubicación
type LocationInfo struct {
	Name          string  `json:"name" bson:"name"`
	Address       string  `json:"address" bson:"address"`
	Time          string  `json:"time" bson:"time"`
	GoogleMapsURL string  `json:"googleMapsUrl" bson:"googleMapsUrl"`
	WazeURL       *string `json:"wazeUrl" bson:"wazeUrl"`
	Notes         *string `json:"notes,omitempty" bson:"notes,omitempty"`
}

// BankAccountInfo contiene los datos para transferencias bancarias o regalos
type BankAccountInfo struct {
	Bank          string  `json:"bank" bson:"bank"`
	AccountHolder string  `json:"accountHolder" bson:"accountHolder"`
	CLABE         string  `json:"clabe" bson:"clabe"`
	AccountNumber *string `json:"accountNumber,omitempty" bson:"accountNumber,omitempty"`
}

// GiftRegistryStore representa una tienda de mesa de regalos
type GiftRegistryStore struct {
	Name        string  `json:"name" bson:"name"`
	URL         string  `json:"url" bson:"url"`
	EventNumber *string `json:"eventNumber,omitempty" bson:"eventNumber,omitempty"`
}

// GiftRegistryConfig agrupa las opciones de la mesa de regalos
type GiftRegistryConfig struct {
	Message     string              `json:"message" bson:"message"`
	Stores      []GiftRegistryStore `json:"stores,omitempty" bson:"stores,omitempty"`
	BankAccount *BankAccountInfo    `json:"bankAccount,omitempty" bson:"bankAccount,omitempty"`
}

// DressCodeConfig define las pautas de vestimenta del evento
type DressCodeConfig struct {
	Title        string   `json:"title" bson:"title"`
	Description  string   `json:"description" bson:"description"`
	PaletteNotes []string `json:"paletteNotes,omitempty" bson:"paletteNotes,omitempty"`
}

// AudioConfig contiene la configuración de la música de fondo para la invitación
type AudioConfig struct {
	Src   string `json:"src" bson:"src"`
	Title string `json:"title" bson:"title"`
}

// HostsConfig representa la información de los anfitriones del evento
type HostsConfig struct {
	Bride        string   `json:"bride" bson:"bride"`
	Groom        string   `json:"groom" bson:"groom"`
	Title        string   `json:"title" bson:"title"`
	WelcomeWords []string `json:"welcomeWords,omitempty" bson:"welcomeWords,omitempty"`
}

// DateConfig contiene la información de la fecha y hora del evento
type DateConfig struct {
	ISO     string `json:"iso" bson:"iso"`
	Display string `json:"display" bson:"display"`
	Time    string `json:"time" bson:"time"`
}

// LocationsConfig agrupa los lugares de ceremonia y recepción
type LocationsConfig struct {
	Ceremony  LocationInfo `json:"ceremony" bson:"ceremony"`
	Reception LocationInfo `json:"reception" bson:"reception"`
}

// Event representa la entidad completa del evento en la base de datos y la API
type Event struct {
	ID           string              `json:"id" bson:"_id,omitempty"`
	Hosts        HostsConfig         `json:"hosts" bson:"hosts"`
	Date         DateConfig          `json:"date" bson:"date"`
	Locations    LocationsConfig     `json:"locations" bson:"locations"`
	DressCode    *DressCodeConfig    `json:"dressCode,omitempty" bson:"dressCode,omitempty"`
	GiftRegistry *GiftRegistryConfig `json:"giftRegistry,omitempty" bson:"giftRegistry,omitempty"`
	Audio        *AudioConfig        `json:"audio,omitempty" bson:"audio,omitempty"`
}

// EventRepository define las operaciones de persistencia que cualquier almacén de datos (MongoDB) debe implementar
type EventRepository interface {
	// FindAll obtiene todos los eventos registrados en la colección
	FindAll(ctx context.Context) ([]Event, error)

	// FindByID busca un evento específico por su identificador
	FindByID(ctx context.Context, id string) (*Event, error)

	// FindFirst obtiene el primer evento disponible (Útil para la landing pública del evento actual)
	FindFirst(ctx context.Context) (*Event, error)
}
