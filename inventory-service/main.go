package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize database
	if err := InitDB(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// API routes
	api := r.Group("/api/v1/inventory")
	{
		// Products
		api.GET("/products", GetProducts)
		api.PUT("/products/:id/price", requireIdentity(true), UpdateProductPrice)

		// Stock management
		api.GET("/stock", GetStock)
		api.PUT("/stock/:productId/:timeSlot", requireIdentity(true), UpdateStock)
		api.POST("/stock/:productId/:timeSlot/adjust", requireIdentity(true), AdjustStock)

		// Bookings
		api.POST("/bookings", requireIdentity(true), AddBooking)
		api.DELETE("/bookings", requireIdentity(true), RemoveBooking)

		// Notifications
		api.GET("/notifications", requireIdentity(true), GetNotifications)
		api.POST("/notifications", requireIdentity(false), CreateNotification)
		api.PUT("/notifications/:id/status", requireIdentity(true), UpdateNotificationStatus)
		api.DELETE("/notifications/:id", requireIdentity(true), DeleteNotification)

		// Analytics
		api.GET("/analytics/summary", requireIdentity(true), GetAnalyticsSummary)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	log.Printf("Inventory service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
