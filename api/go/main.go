package main

import (
	"log"
	"time"

	_ "image/gif"
	_ "image/jpeg"

	_ "golang.org/x/image/webp"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := InitDatabase(); err != nil {
		log.Printf("Failed to initialize database: %v", err)
		log.Println("Continuing without database functionality...")
	}

	defer func() {
		if DB != nil {
			CloseDatabase()
		}
	}()

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.POST("/auth/register", registerHandler)
	router.POST("/auth/login", loginHandler)
	router.POST("/auth/demo-login", demoLoginHandler)

	auth := router.Group("/auth")
	auth.Use(authMiddleware())
	{
		auth.GET("/me", getMeHandler)
		auth.POST("/change-password", changePasswordHandler)
	}

	router.GET("/palettes", getPalettesHandler)
	router.POST("/palettes", savePaletteHandler)
	router.DELETE("/palettes/:id", deletePaletteHandler)

	router.GET("/workspaces", getWorkspacesHandler)
	router.POST("/workspaces", saveWorkspaceHandler)
	router.DELETE("/workspaces/:id", deleteWorkspaceHandler)
	router.POST("/workspaces/:id/share", shareWorkspaceHandler)
	router.DELETE("/workspaces/:id/share", removeWorkspaceShareHandler)
	router.GET("/shared", getSharedWorkspaceHandler)

	router.POST("/apply-palette", applyPaletteHandler)

	router.GET("/wallhaven/search", wallhavenSearchHandler)
	router.GET("/wallhaven/w/:id", wallhavenGetWallpaperHandler)
	router.GET("/wallhaven/download", wallhavenDownloadHandler)

	log.Println("Starting server on :8088")
	router.Run(":8088")
}
