package main

import (
	"api-gateway/config"
	"api-gateway/handlers"
	"api-gateway/middleware"
	"api-gateway/utils"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	config.LoadConfig()
	if err := utils.ValidateSigningConfiguration(); err != nil {
		logrus.Fatal(err)
	}
	utils.InitRedis()

	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "API Gateway OK"})
	})

	v1 := r.Group("/api/v1")
	{
		// User routes
		users := v1.Group("/users")
		{
			// Public routes (no auth required)
			users.POST("/start-registration", middleware.OTPRateLimitMiddleware(), handlers.ProxyToUserService)
			users.POST("/verify-otp", handlers.ProxyToUserService)
			users.POST("/login-otp", middleware.OTPRateLimitMiddleware(), handlers.ProxyToUserService)
			users.POST("/login-verify-otp", handlers.ProxyToUserService)
			users.POST("/complete-profile", middleware.AuthMiddleware(), handlers.ProxyToUserService)
			users.POST("/login", handlers.ProxyToUserService)

			// Protected routes (auth required)
			protected := users.Group("")
			protected.Use(middleware.AuthMiddleware())
			{
				protected.GET("/profile", handlers.ProxyToUserService)
				protected.PUT("/profile", handlers.ProxyToUserService)
				protected.PUT("/update-profile", handlers.ProxyToUserService)
				protected.PUT("/personal-info", handlers.ProxyToUserService)
				protected.POST("/upload-picture", handlers.ProxyToUserService)
			}
		}

		// Admin routes
		admin := v1.Group("/admin")
		{
			// Public admin routes
			admin.POST("/auth/login", handlers.ProxyToAdminService)

			// Protected admin routes
			admin.Any("/users/*path", handlers.ProxyToAdminService)
			admin.Any("/users", handlers.ProxyToAdminService)
			admin.GET("/me", handlers.ProxyToAdminService)
			admin.POST("/logout", handlers.ProxyToAdminService)
			admin.GET("/activity/recent", handlers.ProxyToAdminService)
			admin.POST("/sessions/:id/revoke", handlers.ProxyToAdminService)
		}

		// Inventory routes
		inventory := v1.Group("/inventory")
		{
			inventory.Any("/*path", handlers.ProxyToInventoryService)
			inventory.Any("", handlers.ProxyToInventoryService)
		}
	}

	// Catch-all for other routes
	r.NoRoute(handlers.RouteBasedOnJWT)

	port := os.Getenv("PORT")
	logrus.Infof("API Gateway starting on port %s", port)
	r.Run(":" + port)
}
