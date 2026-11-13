package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jomir/save-my-date-backend/pkg/response"
)

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedVersion := "1.0.0"
	expectedEnv := "testing"

	// Mock para simular que MongoDB responde exitosamente al ping
	mockPingerSuccess := func(ctx context.Context) error {
		return nil
	}

	handler := NewHandler(expectedVersion, expectedEnv, mockPingerSuccess)

	router := gin.New()
	router.GET("/health", handler.Check)

	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	if err != nil {
		t.Fatalf("Error creando petición: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Esperaba status %d, pero obtuvo %d", http.StatusOK, w.Code)
	}

	var resp struct {
		response.Response
		Data StatusResponse `json:"data"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Error deserializando JSON de respuesta: %v", err)
	}

	if !resp.Success {
		t.Errorf("Esperaba 'success: true', pero obtuvo %v", resp.Success)
	}

	if resp.Data.Status != "healthy" {
		t.Errorf("Esperaba status 'healthy', pero obtuvo %s", resp.Data.Status)
	}

	if resp.Data.Version != expectedVersion {
		t.Errorf("Esperaba version '%s', pero obtuvo '%s'", expectedVersion, resp.Data.Version)
	}

	if resp.Data.Environment != expectedEnv {
		t.Errorf("Esperaba environment '%s', pero obtuvo '%s'", expectedEnv, resp.Data.Environment)
	}

	if resp.Data.Services["api"] != "operational" {
		t.Errorf("Esperaba service api 'operational', pero obtuvo '%s'", resp.Data.Services["api"])
	}

	if resp.Data.Services["database"] != "operational" {
		t.Errorf("Esperaba service database 'operational', pero obtuvo '%s'", resp.Data.Services["database"])
	}
}

func TestHealthCheckDatabaseDegraded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock para simular fallo de conectividad en la base de datos
	mockPingerError := func(ctx context.Context) error {
		return errors.New("timeout de conexión")
	}

	handler := NewHandler("1.0.0", "testing", mockPingerError)

	router := gin.New()
	router.GET("/health", handler.Check)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Esperaba status %d, pero obtuvo %d", http.StatusOK, w.Code)
	}

	var resp struct {
		response.Response
		Data StatusResponse `json:"data"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Error deserializando JSON: %v", err)
	}

	if resp.Data.Services["database"] != "degraded" {
		t.Errorf("Esperaba service database 'degraded', pero obtuvo '%s'", resp.Data.Services["database"])
	}
}
