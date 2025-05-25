package main

import (
	"context"
	"fmt"

	"github.com/nikojunttila/community/db"
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
