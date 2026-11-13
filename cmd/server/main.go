package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/jomir/save-my-date-backend/config"
	_ "github.com/jomir/save-my-date-backend/docs"
	eventHandler "github.com/jomir/save-my-date-backend/internal/handlers/event"
	"github.com/jomir/save-my-date-backend/internal/handlers/health"
	"github.com/jomir/save-my-date-backend/internal/handlers/middleware"
	platformMongo "github.com/jomir/save-my-date-backend/internal/platform/mongodb"
	repoMongo "github.com/jomir/save-my-date-backend/internal/repository/mongodb"
	"github.com/jomir/save-my-date-backend/internal/service"
)

// @title           Save My Date Backend API
// @version         1.1
// @description     API backend robusta, segura y escalable para la plataforma Save My Date.
// @contact.name    Equipo de Desarrollo Save My Date
// @BasePath        /
// @schemes         http https

// setupRouter configura y ensambla todos los middlewares y rutas del servicio.
func setupRouter(cfg *config.Config, mongoClient *platformMongo.Client) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CORS(cfg.AllowedOrigins))
	router.Use(middleware.SecurityHeaders(cfg.Environment == "production"))

	// Wiring condicional de dependencias (permite ejecutar tests unitarios con mongoClient nil)
	var ping health.Pinger
	var evtHandler *eventHandler.Handler

	if mongoClient != nil {
		eventRepo := repoMongo.NewEventRepository(mongoClient.Database())
		eventSvc := service.NewEventService(eventRepo)
		evtHandler = eventHandler.NewHandler(eventSvc)
		ping = mongoClient.Ping
	}

	// Handler de salud con ping a base de datos (o nil en testing)
	healthHdl := health.NewHandler(cfg.AppVersion, cfg.Environment, ping)

	// Redirección amigable desde la raíz a Swagger
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Endpoints públicos de salud
	router.GET("/health", healthHdl.Check)

	// Agrupación de rutas API v1
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthHdl.Check)

		// Rutas para Eventos (si el handler está inicializado)
		if evtHandler != nil {
			eventsGroup := v1.Group("/events")
			{
				eventsGroup.GET("", evtHandler.GetAll)
				eventsGroup.GET("/active", evtHandler.GetActive)
				eventsGroup.GET("/:id", evtHandler.GetByID)
			}
		}
	}

	// Documentación interactiva de Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}

func main() {
	// 1. Cargar Configuración
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error crítico al cargar configuración: %v\n", err)
		os.Exit(1)
	}

	// 2. Configurar Logger Estructurado (JSON en producción, Texto en desarrollo)
	var logHandler slog.Handler
	if cfg.Environment == "production" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	slog.Info("Iniciando servicio...",
		slog.String("app", cfg.AppName),
		slog.String("version", cfg.AppVersion),
		slog.String("env", cfg.Environment),
	)

	// 3. Inicializar conexión a MongoDB con timeout de 10s
	mongoClient, err := platformMongo.NewClient(
		context.Background(),
		cfg.MongoURI,
		cfg.MongoDBName,
		10*time.Second,
	)

	if err != nil {
		slog.Error("Fallo crítico al conectar con MongoDB", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 4. Inicializar Router con las dependencias conectadas
	router := setupRouter(cfg, mongoClient)

	// 5. Configurar Servidor HTTP con Timeouts defensivos
	serverAddr := ":" + cfg.Port
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 6. Inciar Servidor en una Goroutine (no bloqueante)
	go func() {
		slog.Info(fmt.Sprintf("Servidor escuchando en http://localhost%s", serverAddr))
		slog.Info(fmt.Sprintf("Documentación Swagger disponible en http://localhost%s/swagger/index.html", serverAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Fallo crítico en el servidor HTTP", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// 6. Esperar señales del sistema para Apagado Elegante (Graceful Shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Señal de apagado recibida, cerrando conexiones de forma ordenada...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Primero cerramos el Servidor HTTP (deja de aceptar nuevas peticiones)
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Servidor forzado a terminar", slog.String("error", err.Error()))
	}

	// Luego, cerramos el pool de conexiones de MongoDB
	if err := mongoClient.Close(shutdownCtx); err != nil {
		slog.Error("Error al cerrar cliente MongoDB", slog.String("error", err.Error()))
	}

	slog.Info("Servidor apagado de manera segura. Ahi nos olemos")
}
