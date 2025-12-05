package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gagehenrich/bucket/internal/config"
	"github.com/gagehenrich/bucket/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type uploadRequest struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size_bytes"`
}

func RequestUpload(cfg *config.Config, pool *pgxpool.Pool, store storage.Storage) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Auth locals set by auth.Middleware
		userIDVal := c.Locals("user_id")
		tierVal := c.Locals("user_tier")

		userID, ok := userIDVal.(string)
		if !ok || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		tier, _ := tierVal.(string)
		if tier == "" {
			tier = "free"
		}

		var req uploadRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}
		if req.Filename == "" || req.Size <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "filename and size required"})
		}

		// Determine retention based on tier
		retention := 72 * time.Hour
		if strings.ToLower(tier) == "premium" {
			retention = 7 * 24 * time.Hour
		}
		expiresAt := time.Now().UTC().Add(retention)

		fileID := uuid.New().String()

		objectPath := fmt.Sprintf("uploads/%s/%s/%s", userID, fileID, req.Filename)

		// Presigned PUT URL (short lifetime, independent from retention)
		uploadURL, err := store.GenerateUploadURL(objectPath, 15*time.Minute)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "storage error"})
		}

		// Tiny code
		tinyCode, err := randomString(9)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
		}

		// Download secret (user-facing) + hash (DB)
		downloadSecret, err := randomString(16)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
		}
		secretHash := sha256.Sum256([]byte(downloadSecret))
		secretHashHex := hex.EncodeToString(secretHash[:])

		// Insert file row
		_, err = pool.Exec(
			context.Background(),
			`INSERT INTO files
			 (id, user_id, filename, size_bytes, storage_path, tiny_code, download_secret_hash, expires_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			fileID, userID, req.Filename, req.Size, objectPath, tinyCode, secretHashHex, expiresAt,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db error"})
		}

		//base := strings.TrimRight(cfg.BaseURL, "/")
		//tinyURL := fmt.Sprintf("%s/d/%s", base, tinyCode)

return c.JSON(fiber.Map{
    "file_id":    fileID,
    "upload_url": uploadURL,
    "tiny_code":  tinyCode,          
    "secret":     downloadSecret,     
    "expires_at": expiresAt.Format(time.RFC3339Nano),
})
	}
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// hex encodes to 2*n chars; trim to requested size
	return hex.EncodeToString(b)[:n], nil
}
