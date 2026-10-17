// bootstrap-admin creates the explicitly configured development account.
package main

import (
	"admin-service/database"
	"admin-service/utils"
	"log"
	"os"
)

func main() {
	if os.Getenv("APP_ENV") != "development" && os.Getenv("APP_ENV") != "test" {
		log.Fatal("dummy admin bootstrap is allowed only in development/test")
	}
	email, password := os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASSWORD")
	if email == "" || len(password) < 12 {
		log.Fatal("set ADMIN_EMAIL and an ADMIN_PASSWORD of at least 12 characters")
	}
	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	hash, err := utils.HashPassword(password)
	if err != nil {
		log.Fatal(err)
	}
	_, err = database.DB.Exec(`INSERT INTO admins (email,password_hash,name,role) VALUES ($1,$2,'Demo Admin','super_admin') ON CONFLICT (email) DO NOTHING`, email, hash)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Development admin bootstrap complete; existing passwords were not changed")
}
