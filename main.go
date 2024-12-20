package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// URL struct defines the schema for MongoDB documents and request body
type URL struct {
	ID       string    `bson:"_id,omitempty" json:"id"`
	LongURL  string    `bson:"long_url" json:"long_url" binding:"required"`
	ShortURL string    `bson:"short_url" json:"short_url"`
	Exp      *int64    `bson:"exp,omitempty" json:"exp,omitempty"` // Expiration timestamp (optional)
}

var urlCollection *mongo.Collection
var validToken = ""
var mongoURI = ""

func init() {
	validToken = os.Getenv("API_TOKEN")
	if validToken == "" {
		log.Fatal("API_TOKEN is not set in environment variables")
	}

	mongoURI = os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI is not set in environment variables")
	}
}

// Initialize MongoDB connection
func initMongoDB() *mongo.Collection {
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB!")
	return client.Database("urlshortener").Collection("urls")
}

// Generate a random short URL
func generateShortURL() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Create a new random generator with a unique source
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	shortURL := make([]byte, 6) // Generate a 6-character random string
	for i := range shortURL {
		shortURL[i] = letters[rng.Intn(len(letters))]
	}
	return string(shortURL)
}

func main() {
	// Initialize MongoDB collection
	urlCollection = initMongoDB()

	// Initialize Gin router
	r := gin.Default()

	// Endpoint to create or overwrite a short URL
	r.POST("/shorten", func(c *gin.Context) {
		var request URL

		token := c.GetHeader("Authorization")
		if strings.TrimSpace(token) != "Bearer "+validToken {
			log.Printf("Unauthorized request: Invalid or missing token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
			return
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			log.Printf("Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Use a custom short URL if provided, otherwise generate a new one
		shortURL := request.ShortURL
		if shortURL == "" {
			shortURL = generateShortURL()
		}

		// Handle expiration time (if provided)
		var exp *int64
		if request.Exp != nil {
			now := time.Now().Unix()
			expTime := now + (*request.Exp * 60) // Convert minutes to seconds
			exp = &expTime
		}

		// Update or insert the record in MongoDB
		filter := bson.M{"short_url": shortURL}
		update := bson.M{
			"$set": bson.M{
				"long_url":  request.LongURL,
				"short_url": shortURL,
				"exp":       exp,
			},
		}
		opts := options.Update().SetUpsert(true) // Upsert: update if exists, insert if not
		_, err := urlCollection.UpdateOne(context.TODO(), filter, update, opts)
		if err != nil {
			log.Printf("Failed to save URL to MongoDB: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL"})
			return
		}

		// Respond with the short URL and optional expiration time
		c.JSON(http.StatusOK, gin.H{
			"long_url":  request.LongURL,
			"short_url": shortURL,
			"exp":       exp,
		})
	})

	// Endpoint to redirect using the short URL
	r.GET("/:shortURL", func(c *gin.Context) {
		shortURL := c.Param("shortURL")

		// Query MongoDB for the corresponding long URL
		var result URL
		err := urlCollection.FindOne(context.TODO(), bson.M{"short_url": shortURL}).Decode(&result)
		if err != nil {
			log.Printf("URL not found for short URL '%s': %v", shortURL, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}

		// Check if the URL has expired
		if result.Exp != nil && time.Now().Unix() > *result.Exp {
			log.Printf("URL for short URL '%s' has expired", shortURL)
			c.JSON(http.StatusGone, gin.H{"error": "URL has expired"})
			return
		}

		// Redirect to the original long URL
		c.Redirect(http.StatusMovedPermanently, result.LongURL)
	})

	// Start the server on port 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
