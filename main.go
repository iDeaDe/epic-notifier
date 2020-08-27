package main

import (
	"flag"
	"log"
	"os"
	"time"
)

func main() {
	var postCurrent bool
	var silent bool

	flag.BoolVar(&postCurrent, "c", true, "Specify to not post current games.")
	flag.BoolVar(&silent, "s", false, "Specify to post games silently.")
	flag.Parse()

	config := GetConfig("config.yaml")

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	tg := new(TelegramSettings)
	tg.Token = telegramToken
	tg.ChannelName = config.Channel

	for {
		ga := GetGiveaway()

		if postCurrent {
			for _, game := range ga.Games {
				log.Printf("Game: %s\n", game.Title)
				tg.Post(&game, silent)
			}
		} else {
			postCurrent = false
		}

		time.Sleep(time.Until(ga.Next.Add(time.Second * 5)))
	}
}
