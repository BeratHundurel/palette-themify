package main

import (
	"log"

	"image-to-palette/db"
	"image-to-palette/handlers"
	_ "image/gif"
	_ "image/jpeg"

	_ "golang.org/x/image/webp"
)

func main() {
	if err := db.InitDatabase(); err != nil {
		log.Printf("Failed to initialize database: %v", err)
		log.Println("Continuing without database functionality...")
	}

	defer func() {
		if db.DB != nil {
			if err := db.CloseDatabase(); err != nil {
				log.Printf("Failed to close database: %v", err)
			}
		}
	}()

	router := handlers.NewRouter()

	log.Println("Starting server on :8088")
	if err := router.Run(":8088"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
