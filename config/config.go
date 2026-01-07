package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	Port                 string
	DBHost               string
	DBPort               string
	DBName               string
	DBUser               string
	DBPassword           string
	DBRootPassword       string
	PlateRecognizerToken string
}

func LoadEnv() *Env {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	return &Env{
		Port:                 os.Getenv("PORT"),
		DBHost:               os.Getenv("BLUEPRINT_DB_HOST"),
		DBPort:               os.Getenv("BLUEPRINT_DB_PORT"),
		DBName:               os.Getenv("BLUEPRINT_DB_DATABASE"),
		DBUser:               os.Getenv("BLUEPRINT_DB_USERNAME"),
		DBPassword:           os.Getenv("BLUEPRINT_DB_PASSWORD"),
		DBRootPassword:       os.Getenv("BLUEPRINT_DB_ROOT_PASSWORD"),
		PlateRecognizerToken: os.Getenv("PLATE_RECOGNIZER_TOKEN"),
	}
}
