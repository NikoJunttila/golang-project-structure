//Package handlers has all http handlers and utility functions for those
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nikojunttila/community/internal/logger"
	"github.com/rs/zerolog/log"
)

// errResponse represents the structure of an error response in JSON format.
type errResponse struct {
	Error string `json:"error"`
}

// RespondWithError sends a JSON error response with the given status code and logs the error.
// It distinguishes between server-side (5xx) and client-side (4xx) errors for logging.
func RespondWithError(ctx context.Context, w http.ResponseWriter, code int, msg string, err error) {
	if code > 499 {
		logger.Error(ctx, err, msg)
	} else {
		logger.Warn(ctx, err, fmt.Sprintf("Client error response: %s", msg))
	}
	RespondWithJSON(ctx,w, code, errResponse{
		Error: msg,
	})
}

// RespondWithJSON sends a JSON response with the given status code and payload.
// If JSON marshaling fails, it sends a 500 Internal Server Error response.
func RespondWithJSON(ctx context.Context,w http.ResponseWriter, code int, payload any) {
	dat, err := json.Marshal(payload)
	if err != nil {
		logger.Error(ctx, err, fmt.Sprintf("Failed to marshal JSON response %v", payload))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(dat)
	if err != nil {
		log.Warn().Msg("Failed to write JSON response")
	}
}

// MAXSIZE represents 1 megabyte in bytes (1 * 1024 * 1024).
const MAXSIZE = 1048576

// DecodeJSONBody decodes the request body into the given value, limiting the body size.
// It returns true if decoding succeeded, or false if a decoding error occurred (with response already sent).
func DecodeJSONBody(w http.ResponseWriter, r *http.Request, v any, maxSize int64) bool {
	if maxSize == 0 {
		maxSize = MAXSIZE // Default 1MB
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	defer func() {
		if err := r.Body.Close(); err != nil {
			logger.Error(r.Context(), err, "Failed to close request body")
		}
	}()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return DecodeJSONWithError(r.Context(),w,decoder, v)
}

// DecodeJSONWithError decodes JSON using the given decoder into the value `v`.
// It returns true if decoding succeeds, false if an error occurs (and response is sent).
func DecodeJSONWithError(ctx context.Context,w http.ResponseWriter, decoder *json.Decoder, v any) bool {
	if err := decoder.Decode(v); err != nil {
		handleJSONDecodeError(ctx, w, err)
		return false
	}
	return true
}

// handleJSONDecodeError handles specific JSON decoding errors and responds with an appropriate HTTP error.
func handleJSONDecodeError(ctx context.Context,w http.ResponseWriter, err error) {
	switch err := err.(type) {
	case *json.SyntaxError:
		RespondWithError(ctx,w, http.StatusBadRequest, "invalid JSON syntax", err)
	case *json.UnmarshalTypeError:
		RespondWithError(ctx,w, http.StatusBadRequest, fmt.Sprintf("invalid type for field %s", err.Field), err)
	case *http.MaxBytesError:
		RespondWithError(ctx,w, http.StatusRequestEntityTooLarge, "request body too large", err)
	default:
		RespondWithError(ctx,w, http.StatusBadRequest, "error parsing JSON", err)
	}
}

