package middleware

import (
	"plate-recognizer-api/model"
	"plate-recognizer-api/utils"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func AuthMiddleware(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Read from form-data
		username := c.FormValue("username")
		password := c.FormValue("password")

		if username == "" || password == "" {
			return utils.Error(
				c,
				fiber.StatusBadRequest,
				"INVALID_REQUEST",
				"username and password are required",
			)
		}

		var user model.User
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			return utils.Error(
				c,
				fiber.StatusUnauthorized,
				"UNAUTHORIZED",
				"invalid username or password",
			)
		}

		if !user.IsActive {
			return utils.Error(
				c,
				fiber.StatusForbidden,
				"FORBIDDEN",
				"user is inactive",
			)
		}

		if err := bcrypt.CompareHashAndPassword(
			[]byte(user.Password),
			[]byte(password),
		); err != nil {
			return utils.Error(
				c,
				fiber.StatusUnauthorized,
				"UNAUTHORIZED",
				"invalid username or password",
			)
		}

		// Save username to context
		c.Locals("username", user.Username)

		return c.Next()
	}
}
