package main

import (
	"fmt"
	"log"
	"os"

	"plate-recognizer-api/config"
	"plate-recognizer-api/internal/server"
	"plate-recognizer-api/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load environment
	env := config.LoadEnv()

	// Build Postgres DSN
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		env.DBHost,
		env.DBUser,
		env.DBPassword,
		env.DBName,
		env.DBPort,
	)

	// Open DB with SQL logging enabled
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Get native SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql DB: %v", err)
	}

	// Ping DB
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}
	log.Println("database connection OK")

	// ----------------------------------------
	// Optional diagnostics (DB_DIAG=1)
	// ----------------------------------------
	if os.Getenv("DB_DIAG") == "1" {
		log.Println("DB_DIAG enabled, running diagnostics")

		var one int
		if err := sqlDB.QueryRow("SELECT 1").Scan(&one); err != nil {
			log.Printf("DB_DIAG: SELECT 1 failed: %v", err)
		} else {
			log.Printf("DB_DIAG: SELECT 1 OK: %d", one)
		}

		var count int64
		if err := db.Raw(`SELECT count(*) FROM "plate_logs"`).Scan(&count).Error; err != nil {
			log.Printf("DB_DIAG: count plate_logs failed: %v", err)
		} else {
			log.Printf("DB_DIAG: plate_logs rows: %d", count)
		}
	}

	// ----------------------------------------
	// ✅ AutoMigrate (SAFE & RECOMMENDED)
	// ----------------------------------------
	log.Println("running database automigration")

	if err := db.AutoMigrate(
		&model.PlateLog{},
		&model.User{},
	); err != nil {
		log.Fatalf("auto migration failed: %v", err)
	}

	log.Println("database migration completed")

	// ----------------------------------------
	// Start Fiber server
	// ----------------------------------------
	s := server.New(env, db)

	log.Printf("server running on port %s", env.Port)
	log.Fatal(s.App.Listen(":" + env.Port))
}
