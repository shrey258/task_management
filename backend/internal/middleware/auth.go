package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/shrey258/task_management/internal/auth"
)

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		// Check if the Authorization header has the Bearer scheme
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		// Validate the token
		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			status := fiber.StatusUnauthorized
			if err == auth.ErrExpiredToken {
				status = fiber.StatusUnauthorized
			}
			return c.Status(status).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Store user information in the context
		c.Locals("user_id", claims.UserID.Hex()) // Convert ObjectID to string
		c.Locals("email", claims.Email)

		return c.Next()
	}
}
