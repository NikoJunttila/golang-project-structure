package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-chi/jwtauth/v5"
	"github.com/nikojunttila/community/internal/logger"
	"github.com/nikojunttila/community/internal/util"

	"github.com/nikojunttila/community/internal/db"
)

var tokenAuth *jwtauth.JWTAuth

func GetTokenAuth() *jwtauth.JWTAuth {
	return tokenAuth
}

func InitAuth() {
	secret := util.GetEnv("JWT_SECRET")
	tokenAuth = jwtauth.New("HS256", []byte(secret), nil)
	NewAuth()
}

func MakeToken(lookupID string) string {
	_, tokenString, _ := tokenAuth.Encode(map[string]any{"lookupID": lookupID})
	return tokenString
}

var lookupErr = errors.New("Failed to find user with lookupID") 

func GetUserFromContext(ctx context.Context) (db.User, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		logger.Error(err, "Error with getting user from ctx")
		return db.User{}, err
	}
	fmt.Println(claims)
	lookupID, ok := claims["lookupID"].(string)
	if !ok {
		logger.Error(lookupErr,"")
		return db.User{}, lookupErr
	}
	user, err := db.Get().GetUserBylookupID(ctx, lookupID)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Cant find user with lookupID %s", lookupID))
		return db.User{}, err
	}
	return user, nil
}
