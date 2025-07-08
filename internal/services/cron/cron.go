//Package cron contains cron job initializes
package cron

import (
	"context"
	"fmt"

	"github.com/nikojunttila/community/internal/logger"
	"github.com/robfig/cron/v3"
)

func task() {
	fmt.Println("hellope from cronjob")
}
//Setup initializes cron jobs
func Setup() {
	c := cron.New()
	// c.AddFunc("@every 1s", func() {
	// 	fmt.Println("Running data transfer")
	// })
	_, err := c.AddFunc("* * 1 * *", func() {
		task()
	})
	if err != nil {
	logger.Fatal(context.Background(),err,"Failed to add cron job")
  return
	}
	c.Start()
}
