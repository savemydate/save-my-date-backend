package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response representa el formato unificado de respuesta de la API.
type Response struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo detalla el motivo del fallo de una solicitud.
type ErrorInfo struct {
	Code    string        `json:"code" example:"INTERNAL_SERVER_ERROR"`
	Message string        `json:"message" example:"Ha ocurrido un error inesperado."`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail especifica un campo individual y su problema en validaciones.
type ErrorDetail struct {
	Field string `json:"field" example:"email"`
	Issue string `json:"issue" example:"formato de correo inválido"`
}

// Success envía una respuesta HTTP exitosa estándar.
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Data:    data,
	})
}

// OK envía una respuesta HTTP 200 con éxito.
func OK(c *gin.Context, data interface{}) {
	Success(c, http.StatusOK, data)
}

// Created envía una respuesta HTTP 201 con éxito.
func Created(c *gin.Context, data interface{}) {
	Success(c, http.StatusCreated, data)
}

// Error envía una respuesta HTTP con formato unificado de error.
func Error(c *gin.Context, statusCode int, code, message string, details ...ErrorDetail) {
	c.JSON(statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
