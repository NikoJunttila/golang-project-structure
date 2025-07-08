// Package logger contains all logging logic
package logger

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// A single private key for setting and getting the logger from the context.
type contextKey string

const loggerContextKey = contextKey("logger")

// NewContext returns a new context with the provided logger embedded.
// We store a pointer to the logger to avoid issues with older zerolog versions
// where methods had pointer receivers.
func NewContext(ctx context.Context, l *zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, l)
}

// FromContext retrieves the logger from the context.
// If no logger is found, it returns a pointer to a copy of the global logger.
// This is safe to use with both old and new versions of zerolog.
func FromContext(ctx context.Context) *zerolog.Logger {
	if l, ok := ctx.Value(loggerContextKey).(*zerolog.Logger); ok {
		return l
	}
	// Return a pointer to a copy of the global logger to prevent modification of the
	// global instance by upstream changes.
	l := log.Logger
	return &l
}

// Error logs an error with a message. It pulls the logger from the context.
// If the message is empty, it defaults to using the error's message.
func Error(ctx context.Context, err error, message string) {
	if message == "" && err != nil {
		message = err.Error()
	}
	FromContext(ctx).Error().Err(err).Msg(message)
}

// Warn logs a warning message, pulling the logger from the context.
func Warn(ctx context.Context, err error, message string) {
	if err != nil {
		FromContext(ctx).Warn().Err(err).Msg(message)
		return
	}
	FromContext(ctx).Warn().Msg(message)
}

// Info logs an info message, pulling the logger from the context.
func Info(ctx context.Context, message string) {
	FromContext(ctx).Info().Msg(message)
}

// Debug logs a debug message, pulling the logger from the context.
func Debug(ctx context.Context, message string) {
	FromContext(ctx).Debug().Msg(message)
}

// Fatal logs a fatal error and exits, pulling the logger from the context.
func Fatal(ctx context.Context, err error, message string) {
	if message == "" && err != nil {
		message = err.Error()
	}
	FromContext(ctx).Fatal().Err(err).Msg(message)
}

// Setup initializes the global zerolog logger for the application.
// This is useful for logging events that are not part of a request lifecycle.
func Setup() {
	// Create the logs directory if it doesn't exist.
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		_ = os.Mkdir("logs", 0755)
	}

	// Set up lumberjack for log file rotation.
	logFile := &lumberjack.Logger{
		Filename:   "logs/app.log", // Log file name
		MaxSize:    20,             // Max size in MB before rotation
		MaxBackups: 5,              // Max number of old log files to keep
		MaxAge:     28,             // Max age in days to keep a log file
		Compress:   true,           // Compress rotated files
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Also log to console with a pretty format.
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	}

	// Combine both file and console outputs.
	multi := io.MultiWriter(logFile, consoleWriter)

	// Set up the global logger instance.
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()

	// Use the global logger directly for application-level events, as there is no context here.
	log.Info().Msg("Global logger initialized")
}
