package server

import (
	"github.com/gofiber/fiber/v2"

	"plate-recognizer-api/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "plate-recognizer-api",
			AppName:      "plate-recognizer-api",
		}),

		db: database.New(),
	}

	return server
}
