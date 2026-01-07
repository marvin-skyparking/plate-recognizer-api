package handler

import (
	"plate-recognizer-api/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func CreateUserHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateUserRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}

		user, err := service.CreateUser(db, req.Username, req.Password)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message": "user created",
			"user": fiber.Map{
				"id":         user.ID,
				"username":   user.Username,
				"is_active":  user.IsActive,
				"created_at": user.CreatedAt,
				"updated_at": user.UpdatedAt,
			},
		})
	}
}
