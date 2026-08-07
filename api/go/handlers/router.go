package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"themesmith/auth"
	"themesmith/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func NewRouter(appConfig config.Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(otelgin.Middleware(appConfig.ServiceName))
	router.Use(errorTracingMiddleware())
	router.Use(recoveryMiddleware())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     appConfig.AllowedOrigins,
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-Trace-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/health", HealthHandler)
	router.POST("/telemetry/errors", clientErrorRateLimitMiddleware(), ClientErrorHandler)

	router.POST("/auth/register", auth.RegisterHandler)
	router.POST("/auth/login", auth.LoginHandler)
	router.GET("/auth/google", auth.GoogleLoginHandler)
	router.GET("/auth/google/callback", auth.GoogleCallbackHandler)
	router.GET("/auth/google/exchange", auth.GoogleExchangeCodeHandler)
	router.GET("/auth/google/desktop/status", auth.GoogleDesktopStatusHandler)

	authGroup := router.Group("/auth")
	authGroup.Use(auth.AuthMiddleware())
	{
		authGroup.GET("/me", auth.GetMeHandler)
		authGroup.POST("/change-password", auth.ChangePasswordHandler)
		authGroup.GET("/preferences", GetPreferencesHandler)
		authGroup.PUT("/preferences", SavePreferencesHandler)
	}

	router.GET("/palettes", GetPalettesHandler)
	router.POST("/palettes/batch", SavePalettesBatchHandler)
	router.POST("/palettes", SavePaletteHandler)
	router.POST("/palettes/:id/share", SharePaletteHandler)
	router.DELETE("/palettes/:id/share", UnsharePaletteHandler)
	router.DELETE("/palettes/:id", DeletePaletteHandler)
	router.DELETE("/palettes", DeletePalettesBatchHandler)

	router.GET("/themes", GetThemesHandler)
	router.POST("/themes/batch", SaveThemesBatchHandler)
	router.POST("/themes", SaveThemeHandler)
	router.POST("/themes/:id/share", ShareThemeHandler)
	router.DELETE("/themes/:id/share", UnshareThemeHandler)
	router.PUT("/themes/:id", UpdateThemeHandler)
	router.DELETE("/themes/:id", DeleteThemeHandler)
	router.DELETE("/themes", DeleteThemesBatchHandler)
	router.GET("/shared-items", GetSharedItemsHandler)
	router.POST("/apply-palette", ApplyPaletteHandler)

	router.GET("/wallhaven/search", WallhavenSearchHandler)
	router.GET("/wallhaven/w/:id", WallhavenGetWallpaperHandler)
	router.GET("/wallhaven/download", WallhavenDownloadHandler)
	router.GET("/desktop/download", DesktopDownloadHandler)

	return router
}

func errorTracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		if spanContext := span.SpanContext(); spanContext.IsValid() {
			c.Header("X-Trace-ID", spanContext.TraceID().String())
		}

		c.Next()

		if c.Writer.Status() < http.StatusInternalServerError {
			return
		}

		err := errors.New(http.StatusText(c.Writer.Status()))
		if len(c.Errors) > 0 {
			err = c.Errors.Last().Err
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

func recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				err := fmt.Errorf("panic: %v", recovered)
				span := trace.SpanFromContext(c.Request.Context())
				span.RecordError(err, trace.WithAttributes(attribute.String("exception.stacktrace", string(debug.Stack()))))
				span.SetStatus(codes.Error, err.Error())
				log.Printf("Recovered request panic: %v\n%s", recovered, debug.Stack())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
		}()
		c.Next()
	}
}
