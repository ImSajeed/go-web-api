package controllers_test

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"test_project/controllers"
	"test_project/models"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// MockDB is a mock implementation of the *sql.DB interface
type MockDB struct{}

func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	// Implement the mock query method
	// Return a mock result for testing purposes
	return nil, nil
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	// Implement the mock exec method
	// Return a mock result for testing purposes
	return nil, nil
}

// Create a new *sql.DB instance and assign it to the mockDBConnection variable
var mockdb *sql.DB

func mockDBConnection() {
	// Database connection string
	connStr := "postgres://postgres:mysecretpassword@localhost/test?sslmode=disable"

	// Connect to the database
	var err error
	mockdb, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Test the database connection
	err = mockdb.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func TestGetDummyData(t *testing.T) {
	// Create a new Fiber app for testing
	mockDBConnection()
	app := fiber.New()

	// Create mock DB and Redis client
	mockRedisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}) // Use the actual redis.Client instance

	// Define the test route
	app.Get("/dummy", func(c *fiber.Ctx) error {
		return controllers.GetDummyData(c, mockdb, mockRedisClient)
	})

	// Create a test request to the /dummy route
	req := httptest.NewRequest(http.MethodGet, "/dummy", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Parse the response body
	var dummyData []models.DummyData
	err = json.NewDecoder(resp.Body).Decode(&dummyData)
	assert.NoError(t, err)

	// Assert the expected behavior
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, dummyData)
}
