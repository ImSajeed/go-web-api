package main

import (
	"context"
	"database/sql"
	"log"
	"test_project/controllers"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

// Database connection
var db *sql.DB

// Redis client
var redisClient *redis.Client

func initDB() {
	// Database connection string
	connStr := "postgres://postgres:mysecretpassword@localhost/test?sslmode=disable"

	// Connect to the database
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Test the database connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func initRedis() {
	// Create a new Redis client
	var err error
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Test the Redis connection
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	initDB()
	initRedis()

	// Create a new Fiber application
	app := fiber.New()

	// Dummy Data API Endpoints
	app.Get("/dummy", func(c *fiber.Ctx) error {
		return controllers.GetDummyData(c, db, redisClient)
	})
	app.Post("/dummy", func(c *fiber.Ctx) error {
		return controllers.CreateDummyData(c, db, redisClient)
	})
	app.Put("/dummy/:id", func(c *fiber.Ctx) error {
		return controllers.UpdateDummyData(c, db, redisClient)
	})
	app.Delete("/dummy/:id", func(c *fiber.Ctx) error {
		return controllers.DeleteDummyData(c, db, redisClient)
	})

	// Slowest Query API Endpoint
	app.Get("/slowest-queries", func(c *fiber.Ctx) error {
		return controllers.GetSlowestQueries(c, db)
	})

	// Start the Fiber application
	log.Fatal(app.Listen(":3000"))
}
