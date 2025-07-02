package cron

import (
	"fmt"
	"github.com/robfig/cron/v3"
)

func task() {
	fmt.Println("hellope")
}

func SetupCron() {
	c := cron.New()
	// c.AddFunc("@every 1s", func() {
	// 	fmt.Println("Running data transfer")
	// })
	c.AddFunc("* * * * *", func() {
		//every minute
		task()
	})
	c.Start()
}
