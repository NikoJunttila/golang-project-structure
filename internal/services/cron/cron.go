package cron

import (
	"fmt"

	"github.com/robfig/cron/v3"
)

func task() {
	fmt.Println("hellope from cronjob")
}

func SetupCron() {
	c := cron.New()
	// c.AddFunc("@every 1s", func() {
	// 	fmt.Println("Running data transfer")
	// })
	c.AddFunc("* * 1 * *", func() {
		task()
	})
	c.Start()
}
