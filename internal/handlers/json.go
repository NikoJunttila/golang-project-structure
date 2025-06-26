package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

type errResponse struct {
	Error string `json:"error"`
}

func RespondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if code > 499 {
		log.Error().Err(err).Int("status", code).Msg(msg)
	} else {
		log.Warn().Err(err).Int("status", code).Msg("Client error response")
	}
	RespondWithJson(w, code, errResponse{
		Error: msg,
	})
}

func RespondWithJson(w http.ResponseWriter, code int, payload any) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Warn().Msgf("Failed to marshal JSON response %v", payload)
		w.WriteHeader(500)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(dat)
	if err != nil {
		log.Warn().Msg("Failed to write JSON response")
	}
}

const MAXSIZE = 1048576

// Decodes json requests and shows errors
func DecodeJSONBody(w http.ResponseWriter, r *http.Request, v any, maxSize int64) bool {
	if maxSize == 0 {
		maxSize = MAXSIZE // Default 1MB
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Warn().Err(err).Msg("Failed to close request body")
		}
	}()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return DecodeJSONWithError(w, decoder, v)
}

// Returns true if decoding was successful, false if an error occurred (and response was sent)
func DecodeJSONWithError(w http.ResponseWriter, decoder *json.Decoder, v any) bool {
	if err := decoder.Decode(v); err != nil {
		handleJSONDecodeError(w, err)
		return false
	}
	return true
}

// handleJSONDecodeError handles different types of JSON decode errors
func handleJSONDecodeError(w http.ResponseWriter, err error) {
	switch err := err.(type) {
	case *json.SyntaxError:
		RespondWithError(w, http.StatusBadRequest, "invalid JSON syntax", err)
	case *json.UnmarshalTypeError:
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("invalid type for field %s", err.Field), err)
	case *http.MaxBytesError:
		RespondWithError(w, http.StatusRequestEntityTooLarge, "request body too large", err)
	default:
		RespondWithError(w, http.StatusBadRequest, "error parsing JSON", err)
	}
}
