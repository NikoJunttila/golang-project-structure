package util

import (
	"os"

	"github.com/rs/zerolog/log"
)

func GetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Panic().Msgf("Failed to find %s", name)
	}
	return value
}
