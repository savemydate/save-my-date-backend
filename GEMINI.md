# Save My Date - Backend Guidelines & Architecture

Este documento define la arquitectura, lineamientos de diseño, estándares de código, seguridad y mejores prácticas para el desarrollo de **Save My Date Backend** en **Go (Golang)**. Debe ser seguido como guía por desarrolladores y agentes de IA al implementar nuevas funcionalidades o refactorizar el código.

---

## 1. Visión General del Proyecto

- **Nombre:** `save-my-date-backend`
- **Lenguaje:** Go (1.23+)
- **Objetivo Inicial:**
  - Endpoint de comprobación de salud (*health check*) público y accesible sin autenticación (`GET /health` y `GET /api/v1/health`).
  - Documentación interactiva de APIs mediante **Swagger (OpenAPI 3.0 / Swaggo)** en `/swagger/index.html`.
  - Sentar las bases de una arquitectura limpia, mantenible, escalable y segura para el crecimiento del servicio.

---

## 2. Arquitectura de Software y Estructura del Proyecto

Se adopta una variación práctica y pragmática de **Clean Architecture / Ports & Adapters**, siguiendo el estándar de facto de la comunidad Go (*Standard Go Project Layout*):

```text
save-my-date-backend/
├── cmd/
│   └── server/
│       └── main.go              # Punto de entrada de la aplicación y wiring de dependencias
├── config/
│   └── config.go                # Carga de variables de entorno con validación estricta
├── docs/                        # Archivos generados automáticamente por Swagger (swaggo)
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/                    # Código privado de la aplicación (no exportable)
│   ├── domain/                  # Entidades de negocio puras, modelos e interfaces de repositorios
│   │   └── event.go
│   ├── handlers/                # Capa de presentación (controladores HTTP / REST)
│   │   ├── event/               # Endpoints REST para Eventos
│   │   ├── health/              # Health check con validación de ping a BD
│   │   └── middleware/          # CORS, Rate Limiting, Logger, Recovery, Security
│   ├── service/                 # Casos de uso / lógica de negocio central
│   │   └── event_service.go
│   ├── repository/              # Acceso a datos y persistencia
│   │   └── mongodb/             # Implementación de repositorios con driver MongoDB v2
│   │       └── event_repository.go
│   └── platform/                # Inicializadores de infraestructura y pools
│       └── mongodb/             # Cliente de conexión, pool y ping para MongoDB Atlas
├── pkg/                         # Paquetes utilitarios reutilizables y exportables (si aplica)
│   └── response/                # Formateo unificado de respuestas HTTP (éxito y errores)
├── .env.example                 # Plantilla de variables de entorno (sin credenciales reales)
├── .gitignore                   # Exclusión de binarios, .env, temporales
├── GEMINI.md                    # Este archivo de lineamientos
├── go.mod                       # Módulo Go y dependencias
├── go.sum                       # Checksums de dependencias
├── Makefile                     # Tareas comunes (run, build, test, swag, lint)
└── README.md                    # Documentación general de onboarding
```

---

## 3. Principios y Buenas Prácticas en Go (Senior Backend)

1. **Gestión de Errores Idiomática:**
   - No usar `panic` para control de flujo normal; reservar solo para fallos catastróficos al iniciar el servidor (`main.go`).
   - Retornar errores explícitos: `if err != nil { return fmt.Errorf("contexto: %w", err) }`.
   - Utilizar wrapping con `%w` para preservar la cadena de errores y permitir `errors.Is` y `errors.As`.

2. **Propagación de Context (`context.Context`):**
   - El primer parámetro de cada función que involucre I/O (base de datos, llamadas HTTP, colas) debe ser `ctx context.Context`.
   - Respetar cancelaciones y timeouts para evitar fugas de goroutines y saturación de conexiones.

3. **Inyección de Dependencias (DI):**
   - Declarar dependencias mediante interfaces en el consumidor (principio de aceptación de interfaces y retorno de structs).
   - Instanciar y conectar dependencias en `cmd/server/main.go`.

4. **Apagado Elegante (*Graceful Shutdown*):**
   - Capturar señales del sistema operativo (`os.Interrupt`, `syscall.SIGTERM`).
   - Permitir que las peticiones en vuelo finalicen con un timeout de gracia (`srv.Shutdown(ctx)`).

5. **Logging Estructurado:**
   - Usar `log/slog` (estándar de Go) o `uber-go/zap` en formato JSON en producción.
   - Incluir campos contextuales: `request_id`, `method`, `path`, `status`, `duration_ms`, `error`.
   - **Regla de oro:** Nunca registrar en logs contraseñas, tokens, tarjetas ni PII sensible.

---

## 4. Estándares de Seguridad Backend

1. **Gestión de Secretos y Configuración:**
   - Cargar configuración exclusivamente desde variables de entorno (`.env` local, inyectadas en despliegue).
   - Mantener un archivo `.env.example` actualizado con valores simulados.
   - Prohibido hardcodear credenciales, API keys o URLs de bases de datos en el código fuente.

2. **CORS (Cross-Origin Resource Sharing):**
   - Configurar orígenes explícitos según el entorno (`ALLOWED_ORIGINS`). Evitar comodines `*` en producción si hay credenciales involucradas.

3. **Cabeceras de Seguridad HTTP:**
   - Implementar cabeceras defensivas básicas:
     - `X-Content-Type-Options: nosniff`
     - `X-Frame-Options: DENY`
     - `Strict-Transport-Security` (HSTS)
     - `Content-Security-Policy` (CSP)
     - `Referrer-Policy: strict-origin-when-cross-origin`

4. **Validación y Sanitización de Entradas:**
   - Validar todos los payloads entrantes (`DTOs`) usando etiquetas de validación (por ejemplo, `go-playground/validator`).
   - Rechazar de inmediato cualquier payload malicioso o mal estructurado con código HTTP 400 y mensaje claro.

5. **Protección contra DoS / Abusos:**
   - Implementar middleware de **Rate Limiting** (basado en IP o token de usuario).
   - Limitar el tamaño máximo del cuerpo de las peticiones (`http.MaxBytesReader`).

---

## 5. Estrategia y Conexión con Bases de Datos

### 5.1. Bases de Datos Relacionales (PostgreSQL / MySQL)

1. **Pool de Conexiones:**
   - Configurar siempre los parámetros del pool según la carga esperada:
     ```go
     db.SetMaxOpenConns(25)
     db.SetMaxIdleConns(10)
     db.SetConnMaxLifetime(15 * time.Minute)
     db.SetConnMaxIdleTime(5 * time.Minute)
     ```
2. **Prevención de Inyecciones SQL:**
   - Usar siempre consultas parametrizadas o generadores de consultas type-safe (ej. `sqlc` o `gorm` con parámetros seguros).
   - Prohibir concatenación de strings para construir queries SQL.
3. **Manejo de Transacciones:**
   - Utilizar funciones o patrones seguros que garanticen `defer tx.Rollback()` y `tx.Commit()` explícito.
4. **Migraciones Versionadas:**
   - Las alteraciones de esquema se gestionan exclusivamente mediante archivos de migración (ej. `golang-migrate` o `goose`), nunca mediante migraciones automáticas en producción.

### 5.2. Bases de Datos No Relacionales y Caché (Redis / MongoDB)

1. **Redis:**
   - Usar Redis para caché de alto rendimiento, rate limiting y sesiones.
   - Definir convenciones de nombres para claves: `<servicio>:<entidad>:<id>` (ej. `smd:session:usr_123`).
   - **Obligatorio:** Asignar siempre un TTL (Time To Live) a los elementos cacheados para evitar saturación de memoria.
   - Manejar degradación elegante: si Redis no responde, la aplicación debe poder continuar operando (fallback a base de datos con registro de advertencia).
2. **MongoDB (si aplica):**
   - Garantizar la creación de índices en campos de búsqueda frecuentes.
   - Aplicar límites (`limit`) y proyecciones en consultas para optimizar transferencia de red y consumo de memoria.

### 5.3. Resiliencia de Conexión:
- Implementar reintentos con backoff exponencial al conectar durante el arranque.
- Integrar verificaciones de conectividad (ping) en el endpoint de **Health Check**.

---

## 6. Diseño de API REST y Formato de Respuestas

### 6.1. Versionamiento y Convenciones
- Prefijo de API: `/api/v1/...`
- Endpoint público de verificación: `/health` (y alias `/api/v1/health`).
- Documentación interactiva: `/swagger/*`.

### 6.2. Estructura Unificada de Respuestas JSON

**Respuesta Exitosa:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "version": "1.0.0",
    "timestamp": "2026-11-05T06:30:00Z"
  }
}
```

**Respuesta de Error:**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Los datos enviados son inválidos.",
    "details": [
      {
        "field": "email",
        "issue": "formato de correo inválido"
      }
    ]
  }
}
```

---

## 7. Documentación con Swagger (OpenAPI)

1. **Herramienta:** Utilizar `swaggo/swag` para generar la especificación OpenAPI a partir de comentarios en el código.
2. **Comentarios Generales (`cmd/server/main.go`):**
   ```go
   // @title Save My Date Backend API
   // @version 1.0
   // @description API backend para el servicio Save My Date.
   // @contact.name Equipo de Desarrollo
   // @BasePath /api/v1
   ```
3. **Comentarios en Handlers:**
   ```go
   // HealthCheckHandler godoc
   // @Summary      Comprobación del estado del servicio
   // @Description  Retorna el estado operativo de la API y sus componentes
   // @Tags         System
   // @Produce      json
   // @Success      200  {object}  response.Response{data=health.HealthStatus}
   // @Router       /health [get]
   ```
4. **Comando de Generación:**
   ```bash
   swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
   ```

---

## 8. Comandos Útiles de Desarrollo

| Acción | Comando |
| :--- | :--- |
| **Inicializar módulo** | `go mod init github.com/jomir/save-my-date-backend` |
| **Instalar Swag CLI** | `go install github.com/swaggo/swag/cmd/swag@latest` |
| **Generar Swagger Docs** | `swag init -g cmd/server/main.go -o docs` |
| **Descargar dependencias** | `go mod tidy` |
| **Ejecutar servidor** | `go run cmd/server/main.go` |
| **Ejecutar pruebas** | `go test -v -race ./...` |
| **Linter de código** | `golangci-lint run` |

---

## 9. Reglas para Agentes de IA y Contribución

- **Consistencia:** Mantener la separación de responsabilidades estricta entre capas (Handler $\to$ Service $\to$ Repository).
- **No inventar dependencias innecesarias:** Favorecer la biblioteca estándar de Go cuando sea suficiente y frameworks consolidados (ej. `Gin`, `Fiber` o `chi`) cuando se requiera productividad en enrutamiento y middlewares.
- **Validación continua:** Tras cualquier cambio estructural de endpoints, regenerar la documentación de Swagger y verificar que compile limpiamente.

---

## 10. Estado Actual y Memoria del Proyecto (Context Snapshot)

> [!NOTE]
> Esta sección almacena el estado consolidado de la aplicación para que cualquier agente mantenga contexto inmediato sin requerir relectura completa de código.

### 10.1. Especificaciones Técnicas Activas
- **Módulo Go:** `github.com/jomir/save-my-date-backend` (Go 1.27.3+)
- **Framework Web:** Gin (`github.com/gin-gonic/gin v1.12.0`)
- **Driver de Base de Datos:** MongoDB Go Driver v2 (`go.mongodb.org/mongo-driver/v2 v2.5.0`)
- **Documentación Swagger:** `swaggo/swag v1.16.6`, `swaggo/gin-swagger v1.6.1`, `swaggo/files v1.0.1`
- **Configuración:** `config/config.go` utilizando `joho/godotenv v1.5.1` y variables de entorno (`.env`).
- **Puerto por defecto:** `:8080`
- **Base de Datos Configurada:** MongoDB (`save_my_date` en MongoDB Atlas / integración Vercel).

### 10.2. Componentes y Módulos Implementados
1. **Punto de Entrada & Wiring (`cmd/server/main.go`):**
   - Inicialización del pool de conexiones a MongoDB con timeout configurable.
   - Inyección de dependencias completa: `platform/mongodb` $\to$ `repository/mongodb` $\to$ `service` $\to$ `handlers`.
   - Función `setupRouter(cfg, mongoClient)` con soporte desacoplado para tests unitarios (permite `mongoClient = nil`).
   - Graceful shutdown ante señales `SIGINT` / `SIGTERM` con cierre ordenado tanto del servidor HTTP como del cliente MongoDB.
   - Logger `log/slog` (JSON en producción, Texto en desarrollo).
2. **Infraestructura de Base de Datos (`internal/platform/mongodb/client.go`):**
   - Cliente singleton con pool de conexiones gestionado (`MaxPoolSize: 50`, `MinPoolSize: 5`, `MaxConnIdleTime: 5m`).
   - Validación de conectividad en arranque mediante `Ping` primario con timeout defensivo.
   - Método `Close(ctx)` para desconexión limpia sin pérdida de transacciones en vuelo.
3. **Dominio de Eventos (`internal/domain/event.go`):**
   - Structs tipados completos: `Event`, `CoupleConfig`, `DateConfig`, `LocationsConfig`, `LocationInfo`, `DressCodeConfig`, `GiftRegistryConfig`, `BankAccountInfo`, `AudioConfig`.
   - Etiquetas duales `bson` y `json` con soporte para campos opcionales (`omitempty`).
   - Definición del contrato `EventRepository` (`FindAll`, `FindByID`, `FindFirst`).
4. **Repositorio MongoDB (`internal/repository/mongodb/event_repository.go`):**
   - Implementa `domain.EventRepository` sobre la colección `events`.
   - Decodificación BSON resiliente con compatibilidad dual para identificadores: soporta `_id` como `bson.ObjectID` hexadecimal nativo de MongoDB Atlas o como `string` personalizado sin errores de casteo.
   - Consultas con filtros BSON dinámicos y manejo de cursores (`cursor.All`) con cierre garantizado (`defer cursor.Close`).
5. **Capa de Servicio de Negocio (`internal/service/event_service.go`):**
   - Casos de uso: `GetAll`, `GetByID`, `GetActiveEvent`.
   - Errores centinela de dominio: `ErrEventNotFound` y `ErrInvalidID`.
   - Validación de parámetros de entrada antes de consultar la persistencia.
6. **Capa de Controladores HTTP (`internal/handlers/event/handler.go`):**
   - Endpoints REST implementados:
     - `GET /api/v1/events`: Listado completo de eventos.
     - `GET /api/v1/events/active`: Consulta del evento principal para renderizar la invitación web.
     - `GET /api/v1/events/:id`: Detalle de evento por ID o Mongo ObjectId.
   - Formateo estandarizado mediante `pkg/response` (`response.OK`, `response.Error`).
   - Anotaciones Swaggo para autogeneración de OpenAPI.
7. **Health Check con Ping a Base de Datos (`internal/handlers/health/health.go`):**
   - Monitoreo en vivo de conectividad con MongoDB.
   - Reporta estado de dependencias: `"services": {"api": "operational", "database": "operational"|"degraded"}`.
   - Endpoints: `GET /health` y `GET /api/v1/health`.
8. **Documentación Swagger (`docs/`):**
   - UI interactiva en `GET /swagger/index.html`.
   - Redirección automática de la raíz `GET /` $\to$ `/swagger/index.html`.
9. **Pruebas Automatizadas:**
   - `internal/handlers/health/health_test.go`: Pruebas de salud con mocks para estados operacional y degradado de base de datos.
   - `cmd/server/main_test.go`: Valida integración de endpoints y middlewares sin requerir base de datos física activa.

### 10.3. Próximos Pasos en el Roadmap
- Integración de la colección de confirmaciones de asistencia (`rsvps`):
  - Modelo de dominio `RSVP` en `internal/domain/rsvp.go`.
  - Endpoint de registro de confirmación `POST /api/v1/rsvp`.
  - Endpoints de consulta y filtrado para el panel de administración `GET /api/v1/rsvps` y métricas `GET /api/v1/rsvps/stats`.
- Autenticación y protección de endpoints administrativos (JWT o API Key).

