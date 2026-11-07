# Save My Date - Backend

API backend robusta, segura y escalable construida en **Go (Golang)** siguiendo principios de **Clean Architecture** y las mejores prácticas de ingeniería de software para servicios backend.

---

## 🚀 Características Iniciales

- **Arquitectura Limpia & Modular:** Separación estricta de responsabilidades (Handlers, Middlewares, Services, Repositories).
- **Health Check Público:** Endpoint para supervisión de salud y disponibilidad (`GET /health` y `GET /api/v1/health`).
- **Documentación Swagger / OpenAPI 3.0:** Generación interactiva y visualizable en tiempo de ejecución en `/swagger/index.html`.
- **Seguridad Defensiva (OWASP):**
  - Cabeceras de seguridad HTTP (`X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `CSP`, `HSTS`).
  - CORS configurable con control de orígenes.
  - Trazabilidad y correlación con `X-Request-ID`.
- **Logging Estructurado:** Uso de `log/slog` con soporte para formato JSON en producción y texto legible en desarrollo.
- **Apagado Elegante (*Graceful Shutdown*):** Cierre seguro de conexiones activas ante señales `SIGINT` / `SIGTERM`.

---

## 📋 Requisitos Previos

- [Go](https://go.dev/dl/) 1.23+ instalado.
- [Swag CLI](https://github.com/swaggo/swag): Para regenerar la documentación OpenAPI (`go install github.com/swaggo/swag/cmd/swag@latest`).

---

## 🛠️ Instalación y Configuración

1. **Clonar el repositorio:**
   ```bash
   git clone https://github.com/jomir/save-my-date-backend.git
   cd save-my-date-backend
   ```

2. **Configurar variables de entorno:**
   Copia el archivo `.env.example` a `.env`:
   ```bash
   cp .env.example .env
   ```

3. **Descargar dependencias:**
   ```bash
   go mod tidy
   ```

---

## 🏃 Ejecución

### Modo Desarrollo:
```bash
go run cmd/server/main.go
```

El servidor iniciará en el puerto configurado (por defecto `http://localhost:8080`).

---

## 📖 Endpoints Disponibles

| Método | Ruta | Descripción | Acceso |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | Redirección a la documentación de Swagger | Público |
| `GET` | `/health` | Health check del sistema | Público |
| `GET` | `/api/v1/health` | Health check versionado del sistema | Público |
| `GET` | `/swagger/*any` | Interfaz interactiva de Swagger UI | Público |

### Ejemplo de respuesta (`GET /health`):
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "version": "1.0.0",
    "environment": "development",
    "timestamp": "2026-11-05T06:30:00Z",
    "uptime": "0h 15m 22s",
    "services": {
      "api": "operational"
    }
  }
}
```

---

## 📚 Documentación con Swagger

Para regenerar la documentación OpenAPI tras modificar endpoints o DTOs:

```bash
swag init -g cmd/server/main.go -o docs
```

Acceso al panel interactivo:
👉 **[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)**

---

## 🧪 Pruebas Automatizadas

Ejecutar suite de tests unitarios:
```bash
go test -v -race ./...
```
