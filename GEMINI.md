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
│   ├── handlers/                # Capa de presentación (controladores HTTP / REST)
│   │   ├── health/
│   │   └── middleware/          # CORS, Rate Limiting, Logger, Recovery, Auth
│   ├── service/                 # Casos de uso / lógica de negocio central
│   ├── repository/              # Acceso a datos (PostgreSQL, MySQL, MongoDB, Redis)
│   │   ├── postgres/
│   │   └── redis/
│   └── platform/                # Inicializadores de infraestructura (DB pool, logger, router)
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
- **Documentación Swagger:** `swaggo/swag v1.16.6`, `swaggo/gin-swagger v1.6.1`, `swaggo/files v1.0.1`
- **Configuración:** `config/config.go` utilizando `joho/godotenv v1.5.1` y variables de entorno (`.env`).
- **Puerto por defecto:** `:8080`

### 10.2. Componentes y Módulos Implementados
1. **Punto de Entrada (`cmd/server/main.go`):**
   - Función desacoplada `setupRouter(cfg)` para pruebas y servidor.
   - Graceful shutdown ante señales `SIGINT` / `SIGTERM` con timeout de 10s.
   - Configuración de `http.Server` con timeouts defensivos (Read: 15s, Write: 15s, Idle: 60s).
   - Logger `log/slog` (JSON en producción, Texto en desarrollo).
2. **Middlewares (`internal/handlers/middleware/`):**
   - `CORS`: Orígenes parametrizables (`ALLOWED_ORIGINS`), credenciales, cabeceras y pre-flights.
   - `SecurityHeaders`: Cabeceras OWASP (`nosniff`, `DENY`, `strict-origin-when-cross-origin`, `HSTS` en prod). Excluye CSP restrictivo de `/swagger/*` para renderizado correcto.
   - `RequestLogger`: Inyección/propagación de `X-Request-ID`, cálculo de latencia en milisegundos y logging estructurado.
3. **Manejador de Respuestas (`pkg/response/response.go`):**
   - Estructura unificada: `Response{Success, Data, Error}`.
   - Detalle de errores con código y lista de campos inválidos (`ErrorDetail`).
4. **Health Check (`internal/handlers/health/health.go`):**
   - Retorna: `status`, `version`, `environment`, `timestamp`, `uptime`, y mapa `services`.
   - Expuesto en: `GET /health` y `GET /api/v1/health`.
5. **Documentación Swagger (`docs/`):**
   - Spec OpenAPI en `docs/swagger.json` y `docs/swagger.yaml`.
   - UI interactiva accesible en `GET /swagger/index.html`.
   - Redirección automática de la raíz `GET /` $\to$ `/swagger/index.html`.
6. **Pruebas Automatizadas:**
   - `internal/handlers/health/health_test.go`: Valida DTO de salud y estado HTTP 200.
   - `cmd/server/main_test.go`: Valida integración de todas las rutas (`/`, `/health`, `/api/v1/health`, `/swagger/index.html`, `/swagger/doc.json`).

### 10.3. Próximos Pasos en el Roadmap
- Definición de entidades de dominio en `internal/domain/`.
- Conexión a base de datos relacional (PostgreSQL) con pool de conexiones y migraciones.
- Conexión a Redis para caché y rate limiting distribuido.
- Implementación de casos de uso en `internal/service/` y persistencia en `internal/repository/`.

