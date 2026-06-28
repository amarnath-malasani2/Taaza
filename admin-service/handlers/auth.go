package handlers

import (
	"admin-service/database"
	"admin-service/models"
	"admin-service/utils"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

var validate = validator.New()

// AdminLogin handles admin login with email and password
func AdminLogin(c *gin.Context) {
	var req models.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide valid email and password"})
		return
	}

	// Get admin from database
	var admin models.Admin
	err := database.DB.QueryRow(
		"SELECT id, email, password_hash, name, role, is_active FROM admins WHERE email = $1",
		req.Email,
	).Scan(&admin.ID, &admin.Email, &admin.Password, &admin.Name, &admin.Role, &admin.IsActive)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Check if admin is active
	if !admin.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is disabled. Contact super admin."})
		return
	}

	// Verify password
	logrus.Infof("Attempting password verification for admin: %s", admin.Email)
	logrus.Infof("Password hash from DB: %s", admin.Password[:20]+"...")
	if !utils.CheckPasswordHash(req.Password, admin.Password) {
		logrus.Errorf("Password verification failed for admin: %s", admin.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	logrus.Infof("Password verification successful for admin: %s", admin.Email)

	// Generate JWT token
	token, err := utils.GenerateAdminJWT(admin.ID, admin.Email, admin.Role)
	if err != nil {
		logrus.Error("JWT generation error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Update last login time
	_, err = database.DB.Exec("UPDATE admins SET last_login_at = NOW() WHERE id = $1", admin.ID)
	if err != nil {
		logrus.Error("Failed to update last login:", err)
	}

	// Log admin activity
	utils.LogAdminActivity(admin.ID, "login", "Admin logged in", c.ClientIP())

	// Return response
	c.JSON(http.StatusOK, models.AdminLoginResponse{
		Token: token,
		Admin: models.AdminInfo{
			ID:    admin.ID,
			Email: admin.Email,
			Name:  admin.Name,
			Role:  admin.Role,
		},
	})
}

// GetCurrentAdmin returns the currently logged-in admin's info
func GetCurrentAdmin(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var admin models.Admin
	err := database.DB.QueryRow(
		"SELECT id, email, name, role, is_active, last_login_at, created_at FROM admins WHERE id = $1",
		adminID,
	).Scan(&admin.ID, &admin.Email, &admin.Name, &admin.Role, &admin.IsActive, &admin.LastLoginAt, &admin.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin not found"})
		return
	}
	if err != nil {
		logrus.Error("Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"admin": admin})
}

// AdminLogout logs out the admin (client-side token removal)
func AdminLogout(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if exists {
		utils.LogAdminActivity(adminID.(int), "logout", "Admin logged out", c.ClientIP())
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
