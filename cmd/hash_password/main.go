package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "password123"
	if len(os.Args) > 1 {
		password = os.Args[1]
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Password:", password)
	fmt.Println("Hash:", string(hashedPassword))
}
