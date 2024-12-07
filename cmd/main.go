package main

import (
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/sid-sun/ntfy.tg/cmd/config"
	"github.com/sid-sun/ntfy.tg/pkg/bot"
)

func main() {
	go func() {
		for {
			time.Sleep(time.Minute)
			func() {
				ts := time.Now().Format("20060102-150405")
				os.Mkdir("profiles", 0770)
				fn := "profiles/mem-profile-" + ts + ".pprof"
				file, err := os.Create(fn)
				if err != nil {
					log.Printf("Failed to create mem profile file: %v\n", err)
				}
				defer file.Close()

				if err = pprof.WriteHeapProfile(file); err != nil {
					log.Printf("Failed to write mem profile file: %v\n", err)
				} else {
					log.Printf("profile file written to: %s\n", fn)
				}
			}()
			time.Sleep(29 * time.Minute)
		}
	}()
	cfg := config.Load()
	initLogger(cfg.GetEnv())
	bot.StartBot(cfg, logger)
}
