package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nikojunttila/community/internal/db"
	"github.com/nikojunttila/community/internal/logger"
)

type HealthResponse struct {
	Status string `json:"status"`
	Time   string `json:"timestamp"`
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := HealthResponse{
		Status: "healthy",
		Time:   time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

func HealthCheckDB(w http.ResponseWriter, r *http.Request) {
	start := time.Now() // Add this line

	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	_, err := db.Get().HealthCheck(ctx)
	if err != nil {
		logger.Error(ctx, err, "Database health check problems")
		w.WriteHeader(http.StatusServiceUnavailable)
		// Include error info in response
		json.NewEncoder(w).Encode(map[string]any{
			"status":        "unhealthy",
			"timestamp":     time.Now().UTC().Format(time.RFC3339),
			"response_time": time.Since(start).String(),
			"error":         err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":        "Healthy",
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"response_time": time.Since(start).String(),
	})
}
