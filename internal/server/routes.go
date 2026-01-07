package server

import (
	"plate-recognizer-api/handler"
	"plate-recognizer-api/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// RegisterRoutes sets up all API routes
func (s *FiberServer) RegisterRoutes() {
	// Enable CORS
	s.App.Use(cors.New())

	// ---------------------------
	// Health check routes
	// ---------------------------
	s.App.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "OK"})
	})

	s.App.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	})

	// ---------------------------
	// Plate recognition route
	// ---------------------------
	recognizeHandler := handler.NewRecognizeHandler(
		s.Env.PlateRecognizerToken,
		s.DB,
	)
	// 🔐 Protected route
	s.App.Post(
		"/api/recognize",
		middleware.AuthMiddleware(s.DB),
		recognizeHandler.Recognize,
	)

	// ---------------------------
	// User registration route
	// ---------------------------
	s.App.Post("/api/register", handler.CreateUserHandler(s.DB))
}
