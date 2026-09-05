package main

import (
	"log"
	"mi3ad/internal/config"
	database "mi3ad/internal/db"
	"mi3ad/internal/ws"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	cfg := config.Load()

	gormDB, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

	if err := database.AutoMigrate(gormDB); err != nil {
		log.Fatalf("failed to run auto-migration: %v", err)
	}

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS("*"))

	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	hub := ws.NewHub()
	ws.RegisterRoutes(e, hub)

	// users.RegisterRoutes(e, database)
	// appointments.RegisterRoutes(e, database, hub) // next module, will call hub.Notify() after booking

	log.Printf("mi3ad api listening on :%s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
