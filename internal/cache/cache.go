package cache

import (
	"context"
	"fmt"

	"github.com/hashicorp/golang-lru/v2"
	"github.com/nikojunttila/community/internal/auth"
	"github.com/nikojunttila/community/internal/db"
	"github.com/nikojunttila/community/internal/logger"
)

var userCache *lru.Cache[string, db.User]

func SetupUserCache() {
	var err error
	userCache, err = lru.New[string, db.User](256)
	if err != nil {
		logger.Fatal(context.Background(), err, "Failed to setup cache")
	}
}

func GetUser(ctx context.Context) (db.User, error) {
	lookupID, err := auth.GetUserLookupID(ctx)
	if err != nil {
		return db.User{}, err
	}
	user, ok := userCache.Get(lookupID)
	if ok {
		logger.Info(ctx, "user from cache")
		return user, nil
	} else {
		fmt.Println("no cache found finding user from db")
		user, err := auth.GetUserFromContext(ctx)
		if err != nil {
			return db.User{}, err
		}
		logger.Info(ctx, fmt.Sprintf("Adding user to cache: %s", user.Email))
		userCache.Add(user.LookupID, user)
		return user, nil
	}
}
