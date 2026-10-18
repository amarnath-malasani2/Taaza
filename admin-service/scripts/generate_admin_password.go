package main

import (
	"fmt"
	"os"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := os.Getenv("ADMIN_PASSWORD")
	if len(password)<12 { panic("set ADMIN_PASSWORD to at least 12 characters") }
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hash: %s\n", string(hash))
}
