package handlers

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"themesmith/db"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const maxClientErrorBody = 64 * 1024
const (
	clientErrorLimit       = 120
	clientErrorLimitWindow = time.Minute
	maxRateLimitClients    = 10_000
)

type rateLimitWindow struct {
	count     int
	startedAt time.Time
}

var clientErrorRateLimits = struct {
	sync.Mutex
	clients map[string]rateLimitWindow
}{clients: make(map[string]rateLimitWindow)}

type clientErrorReport struct {
	Message   string `json:"message" binding:"required"`
	Name      string `json:"name"`
	Stack     string `json:"stack"`
	Source    string `json:"source"`
	URL       string `json:"url"`
	UserAgent string `json:"userAgent"`
}

func HealthHandler(c *gin.Context) {
	status := http.StatusServiceUnavailable
	database := "unavailable"
	if db.DB != nil {
		if sqlDB, err := db.DB.DB(); err == nil && sqlDB.PingContext(c.Request.Context()) == nil {
			database = "ok"
			status = http.StatusOK
		}
	}

	c.JSON(status, gin.H{"status": http.StatusText(status), "database": database})
}

func ClientErrorHandler(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxClientErrorBody)

	var report clientErrorReport
	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid error report"})
		return
	}

	message := truncate(strings.TrimSpace(report.Message), 2_000)
	if message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error message is required"})
		return
	}

	_, span := otel.Tracer("themesmith/client-errors").Start(c.Request.Context(), "client.error")
	defer span.End()

	err := errors.New(message)
	span.RecordError(err)
	span.SetStatus(codes.Error, message)
	span.SetAttributes(
		attribute.String("client.error.name", truncate(report.Name, 200)),
		attribute.String("client.error.source", truncate(report.Source, 200)),
		attribute.String("client.error.url", sanitizeReportedURL(report.URL)),
		attribute.String("client.error.user_agent", truncate(report.UserAgent, 500)),
		attribute.String("exception.stacktrace", truncate(report.Stack, 16_000)),
	)

	c.Status(http.StatusAccepted)
}

func clientErrorRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientAddress := c.Request.RemoteAddr
		if host, _, err := net.SplitHostPort(clientAddress); err == nil {
			clientAddress = host
		}

		now := time.Now()
		clientErrorRateLimits.Lock()
		window, exists := clientErrorRateLimits.clients[clientAddress]
		if exists && now.Sub(window.startedAt) >= clientErrorLimitWindow {
			window = rateLimitWindow{}
			exists = false
		}
		if !exists {
			if len(clientErrorRateLimits.clients) >= maxRateLimitClients {
				for address, candidate := range clientErrorRateLimits.clients {
					if now.Sub(candidate.startedAt) >= clientErrorLimitWindow {
						delete(clientErrorRateLimits.clients, address)
					}
				}
			}
			if len(clientErrorRateLimits.clients) >= maxRateLimitClients {
				clientErrorRateLimits.Unlock()
				c.AbortWithStatus(http.StatusTooManyRequests)
				return
			}
			window = rateLimitWindow{startedAt: now}
		}

		window.count++
		clientErrorRateLimits.clients[clientAddress] = window
		clientErrorRateLimits.Unlock()

		if window.count > clientErrorLimit {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
}

func sanitizeReportedURL(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return truncate(parsed.String(), 2_000)
}

func truncate(value string, maxLength int) string {
	if len(value) <= maxLength {
		return value
	}
	return value[:maxLength]
}
