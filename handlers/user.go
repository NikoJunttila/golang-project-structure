package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nikojunttila/community/auth"
	"github.com/nikojunttila/community/db"
)

type createUserParams struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}
type loginParams struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

func PostCreateUserHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := createUserParams{}
	err := decoder.Decode(&params)
	if err != nil {
		RespondWithError(w, 400, fmt.Sprintf("error parsing JSON: %v", err))
		return
	}
	user, err := db.Get().CreateUser(r.Context(), db.CreateUserParams{
		ID:           uuid.New().String(),
		LookupID:     uuid.New().String(),
		Email:        params.Email,
		PasswordHash: params.Password,
		// Name          string
		// AvatarUrl     string
		// Provider      string
		// ProviderID    string
		// EmailVerified bool
		// CreatedAt     time.Time
		// UpdatedAt     time.Time
	})
	if err != nil {
		log.Println("err creating user", err)
		RespondWithError(w, 400, "error creating user")
		return
	}
	fmt.Println(user)
	RespondWithJson(w, 200, "created user")
}
func GetLoginHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := loginParams{}
	err := decoder.Decode(&params)
	if err != nil {
		RespondWithError(w, 400, fmt.Sprintf("error parsing JSON: %v", err))
		return
	}
	if params.Email == "" {
		fmt.Println("no email")
		return
	}
	user, err := db.Get().GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		log.Println("err: ", err)
		RespondWithError(w, 404, "error getting user")
	}
	token := auth.MakeToken(user.LookupID)

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
		// Uncomment below for HTTPS:
		// Secure: true,
		Name:  "jwt", // Must be named "jwt" or else the token cannot be searched for by jwtauth.Verifier.
		Value: token,
	})
	RespondWithJson(w, 200, token)
}

func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	auth.GetUserFromContext(r.Context())

	RespondWithJson(w, 200, "works")
}
