package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/gagehenrich/bucket/internal/auth"
	"github.com/gagehenrich/bucket/internal/config"
	"github.com/gagehenrich/bucket/internal/db"
	"github.com/gagehenrich/bucket/internal/handlers"
	"github.com/gagehenrich/bucket/internal/storage"
)

func main() {
	// Load environment configuration (.env supported)
	cfg := config.Load()

	// Connect to Postgres
	pool := db.Connect(cfg.DBDSN)
	defer pool.Close()

	// Backblaze B2 storage backend
	b2Client, err := storage.NewB2Client(
		cfg.B2Bucket,
		cfg.B2Endpoint,
		cfg.B2Key,  
		cfg.B2Secret,
	)
	if err != nil {
		log.Fatalf("Failed to initialize B2 client: %v", err)
	}

	// Use B2 for all storage operations
	store := b2Client

	// Fiber application
	app := fiber.New()


	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Post("/v1/account/login", handlers.CreateAccount(pool))
	app.Post("/v1/account/keys", handlers.CreateAPIKey(pool))

	v1 := app.Group("/v1", auth.Middleware(pool))

	//v1.Post("/v1/account/login", handlers.AttemptLogin(pool))
	v1.Get("/files", handlers.ListFiles(pool)) 
	v1.Post("/upload/request", handlers.RequestUpload(cfg, pool, store))
	v1.Post("/download/auth", handlers.AuthDownload(pool, store))


	log.Printf("Starting API on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
