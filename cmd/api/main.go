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
	env := config.LoadEnv()

	// Build DSN
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		env.DBHost,
		env.DBUser,
		env.DBPassword,
		env.DBName,
		env.DBPort,
	)
	// Connect to DB (enable Info logging to capture SQL and params)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	// Ensure DB connection is usable (ping)
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql DB from gorm: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("db ping failed: %v", err)
	}

	// Optional diagnostics — set DB_DIAG=1 to run
	if os.Getenv("DB_DIAG") == "1" {
		log.Println("DB_DIAG: running database diagnostics...")

		// Simple driver-level check
		var one int
		if err := sqlDB.QueryRow("SELECT 1").Scan(&one); err != nil {
			log.Printf("DB_DIAG: driver SELECT 1 failed: %v", err)
		} else {
			log.Printf("DB_DIAG: driver SELECT 1 OK: %d", one)
		}

		// GORM Raw with ? placeholder
		var rowsCount int64
		err := db.Raw("SELECT count(*) FROM \"plate_logs\" LIMIT ?", 1).Scan(&rowsCount).Error
		if err != nil {
			log.Printf("DB_DIAG: gorm Raw with ?: %v", err)
		} else {
			log.Printf("DB_DIAG: gorm Raw with ? OK: %d", rowsCount)
		}

		// GORM Raw with $1 placeholder
		err = db.Raw("SELECT count(*) FROM \"plate_logs\" LIMIT $1", 1).Scan(&rowsCount).Error
		if err != nil {
			log.Printf("DB_DIAG: gorm Raw with $1: %v", err)
		} else {
			log.Printf("DB_DIAG: gorm Raw with $1 OK: %d", rowsCount)
		}
	}

	// Safe migration: create tables only if missing to avoid SELECT LIMIT issues
	log.Println(">>> running safe table creation (create if missing)")
	tables := []interface{}{
		&model.PlateLog{},
		&model.User{},
	}
	for _, t := range tables {
		if !db.Migrator().HasTable(t) {
			log.Printf("creating table for %T", t)
			if err := db.Migrator().CreateTable(t); err != nil {
				log.Fatalf("create table failed for %T: %v", t, err)
			}
		} else {
			log.Printf("table already exists for %T, skipping CreateTable", t)
			// For existing tables, ensure new columns exist (add if missing)
			if _, ok := t.(*model.PlateLog); ok {
				if !db.Migrator().HasColumn(&model.PlateLog{}, "ImageURL") {
					log.Println("adding ImageURL column to plate_logs")
					if err := db.Migrator().AddColumn(&model.PlateLog{}, "ImageURL"); err != nil {
						log.Fatalf("failed to add ImageURL column: %v", err)
					}
				}
			}
		}
	}
	// Initialize Fiber server
	s := server.New(env, db)

	log.Printf("Server running on port %s\n", env.Port)
	log.Fatal(s.App.Listen(":" + env.Port))
}
