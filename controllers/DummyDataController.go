package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"test_project/models"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

// Helper function to retrieve dummy data from the database
func getDummyDataFromDB(db *sql.DB) ([]models.DummyData, error) {
	rows, err := db.Query("SELECT id, name FROM dummy_table")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dummyData := []models.DummyData{}
	for rows.Next() {
		var d models.DummyData
		err := rows.Scan(&d.ID, &d.Name)
		if err != nil {
			return nil, err
		}
		dummyData = append(dummyData, d)
	}

	return dummyData, nil
}

// Handler for GET /dummy
func GetDummyData(c *fiber.Ctx, db *sql.DB, redisClient *redis.Client) error {
	// Check if the data is available in the cache
	cacheKey := "dummy_data"
	cachedData, err := redisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		// Cache hit! Return the data from the cache
		return c.SendString(cachedData)
	}

	// Cache miss! Retrieve the data from the database
	dummyData, err := getDummyDataFromDB(db)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	// Cache the data for future use
	dataBytes, err := json.Marshal(dummyData)
	if err != nil {
		log.Println(err) // Handle JSON marshaling error
	} else {
		err = redisClient.Set(context.Background(), cacheKey, dataBytes, time.Hour).Err()
		if err != nil {
			log.Println(err) // Handle cache set error
		}
	}

	// Return the data as JSON
	return c.JSON(dummyData)
}

// Handler for POST /dummy
func CreateDummyData(c *fiber.Ctx, db *sql.DB, redisClient *redis.Client) error {
	d := new(models.DummyData)
	if err := c.BodyParser(d); err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	// Insert the new dummy data into the database
	stmt, err := db.Prepare("INSERT INTO dummy_table (name) VALUES ($1) RETURNING id")
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(d.Name).Scan(&id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	// Invalidate the cache for dummy data
	err = redisClient.Del(context.Background(), "dummy_data").Err()
	if err != nil {
		log.Println(err) // Handle cache invalidation error
	}

	// Return the created dummy data
	d.ID = id
	return c.JSON(d)
}

// Handler for PUT /dummy/:id
func UpdateDummyData(c *fiber.Ctx, db *sql.DB, redisClient *redis.Client) error {
	id := c.Params("id")

	// Check if the dummy data exists in the database
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM dummy_table WHERE id=$1)", id).Scan(&exists)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	if !exists {
		return c.Status(http.StatusNotFound).SendString("Dummy data not found")
	}

	d := new(models.DummyData)
	if err := c.BodyParser(d); err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	// Update the dummy data in the database
	_, err = db.Exec("UPDATE dummy_table SET name=$1 WHERE id=$2", d.Name, id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	// Invalidate the cache for dummy data
	err = redisClient.Del(context.Background(), "dummy_data").Err()
	if err != nil {
		log.Println(err) // Handle cache invalidation error
	}

	return c.SendStatus(http.StatusOK)
}

// Handler for DELETE /dummy/:id
func DeleteDummyData(c *fiber.Ctx, db *sql.DB, redisClient *redis.Client) error {
	id := c.Params("id")

	// Check if the dummy data exists in the database
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM dummy_table WHERE id=$1)", id).Scan(&exists)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	if !exists {
		return c.Status(http.StatusNotFound).SendString("Dummy data not found")
	}

	// Delete the dummy data from the database
	_, err = db.Exec("DELETE FROM dummy_table WHERE id=$1", id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	// Invalidate the cache for dummy data
	err = redisClient.Del(context.Background(), "dummy_data").Err()
	if err != nil {
		log.Println(err) // Handle cache invalidation error
	}

	return c.SendStatus(http.StatusOK)
}

// Handler for GET /slowest-queries
func GetSlowestQueries(c *fiber.Ctx, db *sql.DB) error {
	// Retrieve the slowest queries from the PostgreSQL logs
	rows, err := db.Query("SELECT query, total_plan_time FROM pg_stat_statements ORDER BY total_plan_time DESC")
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}
	defer rows.Close()

	type SlowQuery struct {
		Query     string `json:"query"`
		TotalTime string `json:"total_plan_time"`
	}

	slowestQueries := []SlowQuery{}
	for rows.Next() {
		var q SlowQuery
		err := rows.Scan(&q.Query, &q.TotalTime)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString(err.Error())
		}
		slowestQueries = append(slowestQueries, q)
	}

	return c.JSON(slowestQueries)
}
