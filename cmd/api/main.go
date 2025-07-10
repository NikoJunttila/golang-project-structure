// Package main initializes the whole app
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/nikojunttila/community/internal/auth"
	"github.com/nikojunttila/community/internal/cache"
	"github.com/nikojunttila/community/internal/db"
	"github.com/nikojunttila/community/internal/logger"
	customMW "github.com/nikojunttila/community/internal/middleware"
	"github.com/nikojunttila/community/internal/routes"
	"github.com/nikojunttila/community/internal/services/cron"
	"github.com/nikojunttila/community/internal/services/email"
	"github.com/nikojunttila/community/internal/utility"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// createRequestLogger creates the dedicated logger for requests
func createRequestLogger() zerolog.Logger {
	requestLogFile := &lumberjack.Logger{
		Filename: "logs/requests.log",
		MaxSize:  10,
		Compress: true,
	}
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	}
	multi := io.MultiWriter(requestLogFile, consoleWriter)
	return zerolog.New(multi).With().Timestamp().Logger()
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}

	// Initialize internal services
	cache.SetupUserCache()
	logger.Setup()
	db.InitDefault()
	auth.Setup()
	email.EmailerInit(&email.Mailer)
	cron.Setup()

	// Set up router and middleware
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	requestLogger := createRequestLogger()
	customMW.InitializeMiddleware(r, requestLogger)
	routes.InitializeRoutes(r)

	// Configure and start HTTP server with timeouts
	portAddr := fmt.Sprintf(":%s", utility.GetEnv("PORT"))
	srv := &http.Server{
		Addr:         portAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Info().Msgf("Listening at: %s", portAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("HTTP server error")
	}
}
