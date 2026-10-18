package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"user-service/database"

	"github.com/gin-gonic/gin"
)

// No DB is configured: the handler must reject unauthenticated input before any query.
func TestCompleteProfileRequiresVerifiedIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/complete-profile", CompleteUserProfile)
	req := httptest.NewRequest(http.MethodPost, "/complete-profile", strings.NewReader(`{"email":"victim@example.test","name":"Changed","mobile":"9876543210"}`))
	req.Header.Set("Content-Type", "application/json")
	result := httptest.NewRecorder()
	r.ServeHTTP(result, req)
	if result.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", result.Code)
	}
}

func TestCompleteProfileUsesSessionUserNotSuppliedEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	original := database.DB
	database.DB = db
	defer func() { database.DB = original }()
	mock.ExpectQuery("SELECT id FROM users WHERE mobile").WithArgs("9876543210", 7).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("UPDATE users SET name").WithArgs("New Name", "9876543210", 7).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO user_activity").WillReturnResult(sqlmock.NewResult(1, 1))
	r := gin.New()
	r.POST("/complete-profile", func(c *gin.Context) { c.Set("user_id", 7) }, CompleteUserProfile)
	req := httptest.NewRequest(http.MethodPost, "/complete-profile", strings.NewReader(`{"email":"victim@example.test","name":"New Name","mobile":"9876543210"}`))
	req.Header.Set("Content-Type", "application/json")
	result := httptest.NewRecorder()
	r.ServeHTTP(result, req)
	if result.Code != 200 {
		t.Fatalf("got %d: %s", result.Code, result.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
