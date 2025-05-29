package logger

import (
	// "errors"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

func LoggerSetup() {
	// Set up lumberjack logger for file rotation
	logFile := &lumberjack.Logger{
		Filename:   "logs/app.log", // Log file name
		MaxSize:    20,             // Max size in MB before rotation
		MaxBackups: 5,              // Max number of old log files to keep
		MaxAge:     28,             // Max age in days to keep a log file
		Compress:   true,           // Compress rotated files
	}

	// Optional: Also log to console
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "02-01-2006 15:04:05",
	}
	// Combine both outputs
	multi := io.MultiWriter(logFile, consoleWriter)

	// Set up zerolog with combined output
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()

	// Example usage
	// log.Info().Msg("Application started")

	// err := errors.New("something went wrong")
	// log.Error().Err(err).Msg("Failed operation")
}
