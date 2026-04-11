package handlers

import (
	"time"

	"image-to-palette/auth"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://wails.localhost:9245"},
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.POST("/auth/register", auth.RegisterHandler)
	router.POST("/auth/login", auth.LoginHandler)
	router.GET("/auth/google", auth.GoogleLoginHandler)
	router.GET("/auth/google/callback", auth.GoogleCallbackHandler)
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
	router.POST("/palettes", SavePaletteHandler)
	router.DELETE("/palettes/:id", DeletePaletteHandler)
	router.DELETE("/palettes", DeletePalettesBatchHandler)

	router.GET("/themes", GetThemesHandler)
	router.POST("/themes", SaveThemeHandler)
	router.PUT("/themes/:id", UpdateThemeHandler)
	router.DELETE("/themes/:id", DeleteThemeHandler)
	router.DELETE("/themes", DeleteThemesBatchHandler)
	router.POST("/apply-palette", ApplyPaletteHandler)

	router.GET("/wallhaven/search", WallhavenSearchHandler)
	router.GET("/wallhaven/w/:id", WallhavenGetWallpaperHandler)
	router.GET("/wallhaven/download", WallhavenDownloadHandler)

	return router
}
