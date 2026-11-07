package health

import (
	"encoding/json"
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
	handler := NewHandler(expectedVersion, expectedEnv)

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
}
