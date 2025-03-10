package config

import (
	"os"

	"github.com/joho/godotenv"
)

func LoadENV() {
  godotenv.Load()
}

func GetValue(key string) string {
  val := os.Getenv(key)
  return val
}
