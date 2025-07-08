//Package main used to bootstrap data for local testing
package main

import (
	"context"
	"fmt"

	"github.com/nikojunttila/community/internal/db"
)

func main() {
	ctx := context.Background()

	value, err := db.Get().InsertFoo(ctx, "hellope2")
	if err != nil {
		fmt.Println("err inserting foo", err)
		return
	}
	fmt.Println(value)
	//insert stuff to db for initalization
}
