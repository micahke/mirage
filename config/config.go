package config

import (
	"os"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
)

func LoadENV() {
	godotenv.Load()
}

func GetValue(key string) string {
	val := os.Getenv(key)
	return val
}
