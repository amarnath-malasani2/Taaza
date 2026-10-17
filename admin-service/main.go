package main

import (
	"admin-service/database"
	"admin-service/handlers"
	"admin-service/middleware"
	"admin-service/utils"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := utils.ValidateSigningConfiguration(); err != nil {
		logrus.Fatal(err)
	}
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	if err := database.Connect(); err != nil {
		logrus.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		// Public admin routes (no auth required)
		auth := v1.Group("/admin/auth")
		{
			auth.POST("/login", handlers.AdminLogin)
		}

		// Protected admin routes (auth required)
		admin := v1.Group("/admin")
		admin.Use(middleware.AdminAuthMiddleware())
		{
			admin.GET("/me", handlers.GetCurrentAdmin)
			admin.POST("/logout", handlers.AdminLogout)
			users := admin.Group("/users")
			{
				users.GET("", handlers.GetUsers)
				users.GET("/stats", handlers.GetUserStats)
				users.GET("/:id", handlers.GetUser)
				users.PUT("/:id", handlers.UpdateUser)
				users.GET("/:id/activity", handlers.GetUserActivity)
				users.GET("/:id/addresses", handlers.GetUserAddresses)
				users.GET("/:id/sessions", handlers.GetUserSessions)
			}

			admin.GET("/activity/recent", handlers.GetRecentActivity)
			admin.POST("/sessions/:id/revoke", handlers.RevokeSession)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "admin-service ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	logrus.Infof("Admin service starting on port %s", port)
	r.Run(":" + port)
}
