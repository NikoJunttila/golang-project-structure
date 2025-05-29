package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/nikojunttila/community/internal/auth"
	"github.com/nikojunttila/community/internal/db"
	"github.com/nikojunttila/community/internal/logger"
	customMW "github.com/nikojunttila/community/internal/middleware"
	"github.com/nikojunttila/community/internal/routes"
	"github.com/nikojunttila/community/internal/util"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}
	logger.LoggerSetup()
	db.InitDefault()
	auth.InitAuth()
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	customMW.InitializeMiddleware(r)
	routes.InitializeRoutes(r)
	portAddr := fmt.Sprintf(":%s", util.GetEnv("PORT"))
	fmt.Println("listening at ", portAddr)
	err := http.ListenAndServe(portAddr, r)
	log.Fatalln("wtf??? ", err)
}
