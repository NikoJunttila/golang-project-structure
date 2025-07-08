// Package utility contains utility functions
package utility

import (
	"errors"
	"os"

	"github.com/rs/zerolog/log"
)

var errEnv = errors.New("failed to get from env")

// GetEnv utility function to log fatal if fails to find value from env
func GetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatal().Err(errEnv).Msgf("Failed to find %s", name)
	}
	return value
}
