package handlers

import (
    "time"

    "github.com/gagehenrich/bucket/internal/auth"
    "github.com/gagehenrich/bucket/internal/storage"

    "github.com/gofiber/fiber/v2"
    "github.com/jackc/pgx/v5/pgxpool"
)

type DownloadAuthRequest struct {
    Tiny   string `json:"tiny"`
    Secret string `json:"secret"`
}

func AuthDownload(pool *pgxpool.Pool, store storage.Storage) fiber.Handler {
    return func(c *fiber.Ctx) error {

        var req DownloadAuthRequest

        if err := c.BodyParser(&req); err != nil {
            return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
        }

        var (
            storagePath string
            secretHash  string
            expiresAt   time.Time
            filename    string
        )
        err := pool.QueryRow(
            c.Context(),
            `SELECT storage_path, download_secret_hash, expires_at, filename
            FROM files
            WHERE tiny_code = $1`,
            req.Tiny,
        ).Scan(&storagePath, &secretHash, &expiresAt, &filename)

        if err != nil {
            return c.Status(404).JSON(fiber.Map{"error": "file not found"})
        }

        if time.Now().After(expiresAt) {
            return c.Status(410).JSON(fiber.Map{"error": "file expired"})
        }

        // validate secret
        if !auth.CheckSecret(req.Secret, secretHash) {
            return c.Status(403).JSON(fiber.Map{"error": "invalid secret"})
        }

        // create presigned GET URL
        url, err := store.GenerateDownloadURL(storagePath, 15*time.Minute)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "failed to generate download URL"})
        }

        return c.JSON(fiber.Map{
            "download_url": url,
            "filename":     filename,
        })
    }
}
