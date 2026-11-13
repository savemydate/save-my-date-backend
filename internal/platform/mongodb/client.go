package mongodb

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Client representa la conexion gestionada a MongoDB
type Client struct {
	client   *mongo.Client
	database string
}

// NewClient inicializa el pool de conexiones hacia MongoDB y valida la conectividad inicial con un Ping.
func NewClient(ctx context.Context, uri, databaseName string, timeout time.Duration) (*Client, error) {
	// 1. Crear un contexto con timeout para la conexión y el ping inicial
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	slog.Info("Conectando a MongoDB@Vercel...", slog.String("database", databaseName))

	// 2. Configurar opciones del cliente (Pool de conexiones y timeouts)
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(50).                 // Máximo de conexiones concurrentes en el pool
		SetMinPoolSize(5).                  // Conexiones mínimas siempre listas
		SetMaxConnIdleTime(5 * time.Minute) // Tiempo máximo de inactividad de una conexión

	// 3. Crear el cliente con el driver oficial v2 de MongoDB
	mongoClient, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("Error al inicializar cliente de MongoDB: %w", err)
	}

	// 4. Verificar que realmente hay conectividad y credenciales válidas
	if err := mongoClient.Ping(pingCtx, readpref.Primary()); err != nil {
		// Si el ping falla, cerramos el cliente para no dejar recursos colgados

		_ = mongoClient.Disconnect(context.Background())

		return nil, fmt.Errorf("Error al hacer ping a MongoDB (%s): %w", databaseName, err)
	}

	slog.Info("Conexión exitosa a MongoDB", slog.String("database", databaseName))

	return &Client{
		client:   mongoClient,
		database: databaseName,
	}, nil
}

// Database retorna la referencia a la base de datos configurada
func (c *Client) Database() *mongo.Database {
	return c.client.Database(c.database)
}

// Ping comprueba si la base de datos sigue respondiendo (Útil para el endpoint /health)
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, readpref.Primary())
}

// Close desconecta limpiamente el pool al apagar el servidor (Graceful Shutdown)
func (c *Client) Close(ctx context.Context) error {
	slog.Info("Cerrando conexiones con MongoDB...")
	return c.client.Disconnect(ctx)
}
