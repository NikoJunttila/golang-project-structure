package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nikojunttila/community/db"
	"github.com/nikojunttila/community/types"
)

func CheckUserExists(ctx context.Context, email string) (bool, error) {
	_, err := db.Get().GetUserByEmail(ctx, email)
	if err == nil {
		return true, nil // user exists
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil // user does not exist
	}
	log.Println("error checking user existence:", err)
	return false, err // actual error
}

//check user use
/*	exists, err := auth.CheckUserExists(r.Context(), params.Email)
if err != nil {
	RespondWithError(w, http.StatusInternalServerError, "Internal server error")
	return
}
if exists {
	RespondWithError(w, http.StatusBadRequest, "User already exists")
	return
}*/

// CheckUserExistIfNotCreate checks if a user exists by email, creates if not exists
func CreateUser(ctx context.Context, password string, params types.CreateUserParams, oAuth types.OauthCreate) (db.User, error) {
	var passHash string
	if !oAuth.IsOAuth {
		//use password to hash
		passHash = "hashed" //actual hash here
	}
	createParams := db.CreateUserParams{
		ID:            uuid.New().String(),
		LookupID:      uuid.New().String(),
		Email:         params.Email,
		PasswordHash:  passHash, // OAuth users don't have passwords
		Name:          params.Name,
		AvatarUrl:     params.AvatarUrl,
		Provider:      oAuth.Provider,
		ProviderID:    oAuth.ProviderID,
		EmailVerified: oAuth.EmailVerified,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	newUser, err := db.Get().CreateUser(ctx, createParams)
	if err != nil {
		return db.User{}, err
	}

	return newUser, nil
}
