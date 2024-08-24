package initialize

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadConfig() bool {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		return false
	}
	return true
}
