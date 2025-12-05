package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type accountRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateAccount is idempotent: if the user already exists, it just returns status ok.
func CreateAccount(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req accountRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}
		if req.Email == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email and password required"})
		}

		// Check if user exists
		var existingID string
		var existingHash string 
		err := pool.QueryRow(
			context.Background(),
			`SELECT id FROM users WHERE email = $1`,
			req.Email,
		).Scan(&existingID)
		if err == nil {
			// User already exists; check password
			err := pool.QueryRow(
				context.Background(),
				`SELECT password_hash FROM users WHERE email = $1`,
				req.Email,
			).Scan(&existingHash)
			if err == nil {
				if bcrypt.CompareHashAndPassword([]byte(existingHash), []byte(req.Password)) != nil {
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid password. Please reset password:"})
				}
			}
			return c.JSON(fiber.Map{"status": "ok"})
		}

		// User does not exsit - create new
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
		}

		id := uuid.New().String()
		_, err = pool.Exec(
			context.Background(),
			`INSERT INTO users (id, email, password_hash, tier, created_at)
			 VALUES ($1, $2, $3, 'free', $4)`,
			id, req.Email, string(hash), time.Now().UTC(),
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create account"})
		}

		return c.JSON(fiber.Map{"status": "ok"})
	}
}

// CreateAPIKey verifies email/password and creates a new account key (API key).
// Response: { "api_key": "<uuid>" }
func CreateAPIKey(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req accountRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}
		if req.Email == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email and password required"})
		}

		var userID string
		var passwordHash string

		err := pool.QueryRow(
			context.Background(),
			`SELECT id, password_hash FROM users WHERE email = $1`,
			req.Email,
		).Scan(&userID, &passwordHash)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}

		apiKey := uuid.New().String()
		keyID := uuid.New().String()

		_, err = pool.Exec(
			context.Background(),
			`INSERT INTO api_keys (id, user_id, api_key, created_at)
			 VALUES ($1, $2, $3, $4)`,
			keyID, userID, apiKey, time.Now().UTC(),
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create api key"})
		}

		return c.JSON(fiber.Map{"api_key": apiKey})
	}
}
