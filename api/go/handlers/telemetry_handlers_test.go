package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestClientErrorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/telemetry/errors", ClientErrorHandler)

	t.Run("accepts a valid report", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/telemetry/errors", strings.NewReader(`{
			"message":"Something failed",
			"name":"TypeError",
			"source":"window.error"
		}`))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusAccepted, response.Code)
	})

	t.Run("rejects a missing message", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/telemetry/errors", strings.NewReader(`{"source":"console.error"}`))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusBadRequest, response.Code)
	})
}
