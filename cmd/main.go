package main

import (
	"log"
	"go-shortener/internal/api"
	"go-shortener/internal/database"
	"github.com/gofiber/fiber/v2"
)

func main() {
	if err := database.Connect(); err != nil {
		log.Fatalf("could not connect to databases: %v", err)
	}

	app := fiber.New()

	api.SetupRoutes(app)

	log.Fatal(app.Listen(":3000"))
}