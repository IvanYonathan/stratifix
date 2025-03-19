package main

import (
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Connect to PostgreSQL with GORM
	dsn := "host=localhost user=postgres password= dbname=ticketing_db port=5432 sslmode=disable"
	// Override with environment variable if provided
	if os.Getenv("DATABASE_URL") != "" {
		dsn = os.Getenv("DATABASE_URL")
	}

	// Initialize database connection
	if err := initDatabase(dsn); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Serve static frontend files
	e.Static("/", "./frontend")

	// API routes
	e.GET("/api/events", getEvents)
	e.GET("/api/seats/:eventId", getSeats)
	e.POST("/api/bookings", createBooking)
	e.GET("/ws", handleWebSocket)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Connected to database successfully\n")
	fmt.Printf("Starting server on port %s\n", port)
	fmt.Printf("Access your website at http://localhost:%s\n", port)

	e.Logger.Fatal(e.Start(":" + port))
}
