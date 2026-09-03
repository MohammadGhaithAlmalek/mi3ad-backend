package main

import (
	"log"

	"github.com/echo-backend/MohammadGhaithAlmalek/internal/config"
	"github.com/echo-backend/MohammadGhaithAlmalek/internal/db"
	"github.com/echo-backend/MohammadGhaithAlmalek/internal/ws"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

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
