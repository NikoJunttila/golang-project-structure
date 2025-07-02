package util

import (
	"errors"
	"os"

	"github.com/rs/zerolog/log"
)

var envErr = errors.New("Failed to get from env")

func GetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatal().Err(envErr).Msgf("Failed to find %s", name)
	}
	return value
}
