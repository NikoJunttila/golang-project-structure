package handlers

import (
	"errors"
	"net/http"
)

type exampleParams struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

var errEmailAndPass = errors.New("email or password is missing")

// GetFooHandler foo handler
func GetFooHandler(w http.ResponseWriter, r *http.Request) {
	RespondWithJSON(r.Context(), w, http.StatusOK, "foo")
}

// ExampleHandler basic version of golang web service handler reading form data
func ExampleHandler(w http.ResponseWriter, r *http.Request) {
	var params exampleParams
	if !DecodeJSONBody(w, r, &params, 0) {
		return
	}

	// Validate params
	if params.Email == "" || params.Password == "" {
		RespondWithError(r.Context(), w, 400, "email and password are required", errEmailAndPass)
		return
	}
	RespondWithJSON(r.Context(), w, 200, "example")
}
