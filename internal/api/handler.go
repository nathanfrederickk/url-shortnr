package api

import (
	"fmt"
	"log"
	"time"
	"go-shortener/internal/database"
	"go-shortener/internal/shortener"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetupRoutes defines the API endpoints and assigns them to handlers
func SetupRoutes(app *fiber.App) {
	app.Post("/shorten", shortenURL)
	app.Get("/:shortURL", redirectURL)
}

// shortenURL is the handler for creating a new short URL
func shortenURL(c *fiber.Ctx) error {
	// Parse request body
	var request struct {
		URL string `json:"url"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}
	longURL := request.URL

	countersCollection := database.MongoClient.Database("url_shortener").Collection("counters")
	filter := bson.M{"_id": "url_id"}
	update := bson.M{"$inc": bson.M{"sequence_value": 1}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedCounter struct {
		SequenceValue int64 `bson:"sequence_value"`
	}

	// Split the command into two steps to find the exact error.
	singleResult := countersCollection.FindOneAndUpdate(database.Ctx, filter, update, opts)
	if err := singleResult.Err(); err != nil {
		// This logs the specific error from the database operation itself.
		log.Printf("Error from FindOneAndUpdate: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update counter"})
	}

	err := singleResult.Decode(&updatedCounter)
	if err != nil {
		// Logs the specific error from the decoding step.
		log.Printf("Error decoding counter result: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not decode counter"})
	}

	// Convert the ID to a Base62 string using the 'shortener' package
	shortCode := shortener.ToBase62(updatedCounter.SequenceValue)


	// Create the URL record and save to MongoDB
	urlRecord := shortener.URL{
		ID:        shortCode,
		LongURL:   longURL,
		CreatedAt: time.Now(),
	}
	urlsCollection := database.MongoClient.Database("url_shortener").Collection("urls")
	_, err = urlsCollection.InsertOne(database.Ctx, urlRecord)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not save URL"})
	}

	// Return the response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"short_url": "http://localhost:3000/" + shortCode,
	})
}

// redirectURL handles the redirection from short URL to original URL
func redirectURL(c *fiber.Ctx) error {
	shortURL := c.Params("shortURL")

	// Check Redis Cache first
	longURL, err := database.RedisClient.Get(database.Ctx, shortURL).Result()
	if err == nil {
		fmt.Println("Cache hit for:", shortURL)
		return c.Redirect(longURL, fiber.StatusMovedPermanently)
	}

	// If cache miss, query MongoDB
	fmt.Println("Cache miss for:", shortURL)
	urlsCollection := database.MongoClient.Database("url_shortener").Collection("urls")
	var urlRecord shortener.URL
	err = urlsCollection.FindOne(database.Ctx, bson.M{"_id": shortURL}).Decode(&urlRecord)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "URL not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	// Store the result in Redis for future requests
	err = database.RedisClient.Set(database.Ctx, shortURL, urlRecord.LongURL, 24*time.Hour).Err()
	if err != nil {
		log.Println("Could not set Redis cache:", err) // Non-fatal error
	}

	// Redirect the user
	return c.Redirect(urlRecord.LongURL, fiber.StatusMovedPermanently)
}
