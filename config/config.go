package config

import (
	"os"

  _ "github.com/joho/godotenv/autoload"
)

func LoadENV() {
	// godotenv.Load()
}

func GetValue(key string) string {
	val := os.Getenv(key)
	return val
}
