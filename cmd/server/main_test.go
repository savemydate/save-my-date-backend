package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jomir/save-my-date-backend/config"
)

func TestSetupRouterEndpoints(t *testing.T) {
	cfg := &config.Config{
		Port:           "8080",
		Environment:    "testing",
		AppName:        "save-my-date-test",
		AppVersion:     "1.0.0",
		AllowedOrigins: []string{"*"},
	}

	router := setupRouter(cfg)

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			name:           "Root redirects to Swagger",
			method:         http.MethodGet,
			url:            "/",
			expectedStatus: http.StatusMovedPermanently,
		},
		{
			name:           "Health Check Root",
			method:         http.MethodGet,
			url:            "/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Health Check Versioned API v1",
			method:         http.MethodGet,
			url:            "/api/v1/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Swagger UI Index",
			method:         http.MethodGet,
			url:            "/swagger/index.html",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Swagger Doc JSON",
			method:         http.MethodGet,
			url:            "/swagger/doc.json",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Para %s %s esperaba status %d, pero obtuvo %d", tt.method, tt.url, tt.expectedStatus, w.Code)
			}
		})
	}
}
