package main

import (
	"flag"
	"log"
	"os"
	"path"
	"time"
)

func main() {
	var postCurrent bool
	var silent bool
	var testChannel string

	flag.BoolVar(&postCurrent, "c", true, "Specify to not post current games.")
	flag.BoolVar(&silent, "s", false, "Specify to post games silently.")
	flag.StringVar(&testChannel, "test", "", "Post to the test channel")
	flag.Parse()

	// Для systemd
	executable, err := os.Executable()
	if err != nil {
		log.Panicln(err)
	}
	execPath := path.Dir(executable)
	err = os.Chdir(execPath)
	if err != nil {
		log.Panicln(err)
	}

	config := GetConfig("config.yaml")

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	tg := new(TelegramSettings)
	tg.Token = telegramToken
	tg.ChannelName = config.Channel
	if testChannel != "" {
		tg.ChannelName = testChannel
	}

	for {
		ga := GetGiveaway()
		nextGiveaway := ga.Next

		if time.Now().After(nextGiveaway) && time.Now().Before(time.Now().AddDate(0, 0, -7)) {
			nextGiveaway = time.Now().Add(time.Hour * 24 * 3)
			postCurrent = false
		}

		if postCurrent {
			for _, game := range ga.Games {
				log.Printf("Game: %s\n", game.Title)
				tg.Post(&game, silent)
			}
		} else {
			postCurrent = true
			log.Println("Nothing to post")
		}

		log.Println("Next giveaway:", nextGiveaway.String())
		time.Sleep(time.Until(ga.Next.Add(time.Second * 5)))
	}
}
