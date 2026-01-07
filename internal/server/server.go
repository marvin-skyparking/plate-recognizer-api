package server

import (
	"log"
	"plate-recognizer-api/config"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type FiberServer struct {
	App *fiber.App
	Env *config.Env
	DB  *gorm.DB
}

// New creates a new FiberServer and requires db as argument
func New(env *config.Env, db *gorm.DB) *FiberServer {
	app := fiber.New()

	// Check if db is nil
	if db == nil {
		log.Fatal("database connection is nil")
	}

	server := &FiberServer{
		App: app,
		Env: env,
		DB:  db, // assign DB properly
	}

	server.RegisterRoutes()
	return server
}
