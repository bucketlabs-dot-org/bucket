package handlers

import (
    "time"
    
    "github.com/gofiber/fiber/v2"
    "github.com/jackc/pgx/v5/pgxpool"
)

func ListFiles(pool *pgxpool.Pool) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Locals("user_id").(string)
        
        rows, err := pool.Query(c.Context(), `
            SELECT id, filename, size_bytes, tiny_code, expires_at
            FROM files
            WHERE user_id = $1
            ORDER BY created_at DESC
        `, userID)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "db error"})
        }
        defer rows.Close()
        
        type FileEntry struct {
            ID        string    `json:"id"`
            Filename  string    `json:"filename"`
            Size      int64     `json:"size_bytes"`
            TinyCode  string    `json:"tiny_code"`
            ExpiresAt time.Time `json:"expires_at"`
        }
        
        files := []FileEntry{}
        for rows.Next() {
            var f FileEntry
            if err := rows.Scan(&f.ID, &f.Filename, &f.Size, &f.TinyCode, &f.ExpiresAt); err != nil {
                return c.Status(500).JSON(fiber.Map{"error": "scan error: " + err.Error()})
            }
            files = append(files, f)
        }
        
        if err := rows.Err(); err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "row error: " + err.Error()})
        }
        
        return c.JSON(files)
    }
}