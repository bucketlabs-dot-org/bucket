package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Middleware authenticates requests using the account key (api_keys.api_key).
// It expects: Authorization: Bearer <account-key>
//
// On success it sets:
//   c.Locals("user_id")   -> string (uuid)
//   c.Locals("user_tier") -> string ("free", "premium", ...)
func Middleware(pool *pgxpool.Pool) fiber.Handler {
    return func(c *fiber.Ctx) error {

        // Allow unauthenticated access to account creation/login/key routes
        if strings.HasPrefix(c.Path(), "/v1/account") {
            return c.Next()
        }

        // Everything else requires API key
        token := c.Get("Authorization")
        if token == "" || !strings.HasPrefix(token, "Bearer ") {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error":"missing or invalid authorization header"})
        }

        apiKey := strings.TrimPrefix(token, "Bearer ")

        var userID string
        err := pool.QueryRow(context.Background(),
            `SELECT user_id FROM api_keys WHERE api_key = $1`, apiKey,
        ).Scan(&userID)
        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error":"invalid api key"})
        }

        c.Locals("user_id", userID)
        return c.Next()
    }
}


// HashSecret hashes a plaintext download secret using SHA256.
// Called when a file is created.
func HashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// CheckSecret compares the provided secret against the stored hash.
// Called during download authentication.
func CheckSecret(secret, storedHash string) bool {
	return HashSecret(secret) == storedHash
}
