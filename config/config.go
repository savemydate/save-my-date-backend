package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config almacena la configuración global de la aplicación.
type Config struct {
	Port           string
	Environment    string
	AppName        string
	AppVersion     string
	AllowedOrigins []string
	RateLimitRPS   int
	RateLimitBurst int
	MongoURI       string //URI de la conexion (Comienza por "mongodb+srv://...")
	MongoDBName    string // Nombre de la BD
}

// Load carga las variables de entorno desde .env o el entorno del sistema.
func Load() (*Config, error) {
	// Intenta cargar .env sin fallar si no existe (típico en entornos containerizados)
	if err := godotenv.Load(); err != nil {
		slog.Info("Archivo .env no encontrado o no pudo ser cargado, utilizando variables del sistema")
	}

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		AppName:        getEnv("APP_NAME", "save-my-date-backend"),
		AppVersion:     getEnv("APP_VERSION", "1.0.0"),
		AllowedOrigins: parseSlice(getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173,http://localhost:8080,http://localhost:4200")),
		RateLimitRPS:   getEnvAsInt("RATE_LIMIT_RPS", 100),
		RateLimitBurst: getEnvAsInt("RATE_LIMIT_BURST", 200),
		MongoURI:       getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBName:    getEnv("MONGODB_DATABASE_NAME", "test-project"),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return fallback
}

func parseSlice(val string) []string {
	parts := strings.Split(val, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
