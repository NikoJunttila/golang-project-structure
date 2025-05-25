package util

import (
	"log"
	"os"
)

func GetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Panicln("Failed to find ", name)
	}
	return value
}
