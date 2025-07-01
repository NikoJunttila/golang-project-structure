package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-chi/jwtauth/v5"
	"github.com/nikojunttila/community/internal/db"
	"github.com/nikojunttila/community/internal/logger"
	"github.com/nikojunttila/community/internal/util"
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

// define claim keys to avoid typos
const (
	ClaimLookupID = "lookupID"
	ClaimRole     = "role"
	Admin         = "admin"
	User          = "user"
)

// MakeToken creates a signed JWT for the given user lookupID and optional role
func MakeToken(lookupID string, role ...string) string {
	claims := map[string]any{
		ClaimLookupID: lookupID,
	}
	if len(role) > 0 {
		claims[ClaimRole] = role[0]
	}

	_, tokenString, _ := tokenAuth.Encode(claims)
	return tokenString
}

var ErrLookupIDMissing = errors.New("lookupID not found in token")
var ErrUserNotFound = errors.New("user not found in database")

// GetUserFromContext retrieves the authenticated user from the JWT claims in the request context
func GetUserFromContext(ctx context.Context) (db.User, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		logger.Error(ctx, err, "invalid JWT context")
		return db.User{}, err
	}
	// Log full claims for debugging
	logger.Debug(ctx, fmt.Sprintf("JWT claims %s", claims))

	lookupID, ok := claims[ClaimLookupID].(string)
	if !ok || lookupID == "" {
		logger.Warn(ctx, nil, "lookupID missing from token claims")
		return db.User{}, ErrLookupIDMissing
	}

	user, err := db.Get().GetUserBylookupID(ctx, lookupID)
	if err != nil {
		logger.Error(ctx, err, fmt.Sprintf("no user found for lookupID %s", lookupID))
		return db.User{}, ErrUserNotFound
	}

	return user, nil
}
