package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-chi/jwtauth/v5"
	"github.com/nikojunttila/community/internal/util"

	"github.com/nikojunttila/community/internal/db"

	"github.com/rs/zerolog/log"
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

func MakeToken(name string) string {
	_, tokenString, _ := tokenAuth.Encode(map[string]any{"lookupID": name})
	return tokenString
}
var lookupErr = errors.New("Failed to find user with lookupID") 
func GetUserFromContext(ctx context.Context) (db.User, error) {
	fmt.Println(ctx)
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		log.Warn().Msgf("Error with getting claims %v ", err)
		return db.User{}, err
	}
	lookupID, ok := claims["lookupID"].(string)
	if !ok {
		log.Warn().Msg("Error with getting lookupID")
		return db.User{}, lookupErr
	}
	user, err := db.Get().GetUserBylookupID(ctx, lookupID)
	if err != nil {
		log.Warn().Msgf("Error with finding user with lookupID %v %v", lookupID, err)
	}
	return user, nil
}
