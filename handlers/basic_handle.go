package handlers

import (
	"net/http"
)

type exampleParams struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

func ExampleHandler(w http.ResponseWriter, r *http.Request) {
	var params exampleParams
	if !DecodeJSONBody(w, r, &params, 0) {
		return
	}

	// Validate params
	if params.Email == "" || params.Password == "" {
		RespondWithError(w, 400, "email and password are required")
		return
	}
	RespondWithJson(w, 200, "example")
}
