package handlers

import (
	"net/http"
	"errors"
)

type exampleParams struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}
var emailAndPass = errors.New("email or password is missing")

func GetFooHandler(w http.ResponseWriter, r *http.Request) {
	RespondWithJson(w,r.Context(), http.StatusOK, "foo")
}

func ExampleHandler(w http.ResponseWriter, r *http.Request) {
	var params exampleParams
	if !DecodeJSONBody(w, r, &params, 0) {
		return
	}

	// Validate params
	if params.Email == "" || params.Password == "" {
		RespondWithError(w, r.Context(), 400, "email and password are required", emailAndPass)
		return
	}
	RespondWithJson(w,r.Context(), 200, "example")
}
