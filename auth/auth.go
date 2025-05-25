package auth

import (
	"context"
	"fmt"

	"github.com/go-chi/jwtauth/v5"
	"github.com/nikojunttila/community/util"

	"log"

	"github.com/nikojunttila/community/db"

	"github.com/joho/godotenv"
)

var tokenAuth *jwtauth.JWTAuth

func GetTokenAuth() *jwtauth.JWTAuth {
	return tokenAuth
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}
	secret := util.GetEnv("JWT_SECRET")
	tokenAuth = jwtauth.New("HS256", []byte(secret), nil)
}

func MakeToken(name string) string {
	_, tokenString, _ := tokenAuth.Encode(map[string]any{"username": name})
	return tokenString
}
func GetUserFromContext(ctx context.Context) {
	_, claims, _ := jwtauth.FromContext(ctx)
	email, ok := claims["username"].(string)
	if !ok {
		fmt.Println("username is not a string or is missing")
		return
	}
	fmt.Println(email)

	user, err := db.Get().GetUserByEmail(ctx, email)
	if err != nil {
		log.Println("failed login")
	}
	fmt.Println(user)
}
