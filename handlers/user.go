package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nikojunttila/community/auth"
	"github.com/nikojunttila/community/db"
	"github.com/nikojunttila/community/types"
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
	var params createUserParams
	if !DecodeJSONBody(w, r, &params, 0) {
		return
	}
	// Validation
	if params.Email == "" || params.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	exists, err := auth.CheckUserExists(r.Context(), params.Email)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if exists {
		RespondWithError(w, http.StatusBadRequest, "User already exists")
		return
	}
	cleanedEmail := strings.TrimSpace(strings.ToLower(params.Email))
	createParams := types.CreateUserParams{
		Email: cleanedEmail,
		Name:  "",
	}
	user, err := auth.CreateUser(r.Context(), params.Password, createParams, types.OauthCreate{})
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("error creating user: %v", err))
	}
	fmt.Println(user)
	RespondWithJson(w, http.StatusOK, "created user")
}

func PostLoginHandler(w http.ResponseWriter, r *http.Request) {
	params := loginParams{}
	if !DecodeJSONBody(w, r, &params, 0) {
		return
	}
	if params.Email == "" || params.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "no email or password")
		return
	}
	user, err := db.Get().GetUserByEmail(r.Context(), params.Email)
	if err == sql.ErrNoRows {
		log.Println("err: ", err)
		RespondWithError(w, http.StatusBadRequest, "this email does not exist")
		return
	}
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("database error %v", err))
		return
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
	RespondWithJson(w, http.StatusOK, token)
}

func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	auth.GetUserFromContext(r.Context())

	RespondWithJson(w, 200, "works")
}
