package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Println("Responding with 5XX error:", msg)
	}

	type errResponse struct {
		Error string `json:"error"`
	}

	RespondWithJson(w, code, errResponse{
		Error: msg,
	})
}

func RespondWithJson(w http.ResponseWriter, code int, payload any) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("failed to marshal JSON response %v", payload)
		w.WriteHeader(500)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

const MAXSIZE = 1048576

// Decodes json requests and shows errors
func DecodeJSONBody(w http.ResponseWriter, r *http.Request, v any, maxSize int64) bool {
	if maxSize == 0 {
		maxSize = MAXSIZE // Default 1MB
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	defer r.Body.Close()

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
		RespondWithError(w, http.StatusBadRequest, "invalid JSON syntax")
	case *json.UnmarshalTypeError:
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("invalid type for field %s", err.Field))
	case *http.MaxBytesError:
		RespondWithError(w, http.StatusRequestEntityTooLarge, "request body too large")
	default:
		RespondWithError(w, http.StatusBadRequest, "error parsing JSON")
	}
}
