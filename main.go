package main

import (
	"flag"
	"log"
	"os"
	"path"
	"runtime"
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

	/*
		Смена директории для того, чтобы конфиг создавался рядом с исполняемым файлом
	*/
	executable, err := os.Executable()
	if err != nil {
		log.Panicln(err)
	}
	execPath := path.Dir(executable)
	err = os.Chdir(execPath)
	if err != nil {
		log.Panicln(err)
	}

	config := GetConfig("config.json")

	telegramToken := os.Getenv("TELEGRAM_TOKEN")

	if telegramToken == "" {
		log.Panicln("TELEGRAM_TOKEN not found")
	}

	tg := new(TelegramSettings)
	tg.Token = telegramToken
	tg.ChannelName = config.Channel
	if testChannel != "" {
		tg.ChannelName = testChannel
	}

	for {
		ga := new(Giveaway)
		ga = GetGiveaway()
		nextGiveaway := ga.Next

		if nextGiveaway.Before(time.Now()) {
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

		ga = nil
		runtime.GC()

		log.Println("Next giveaway:", nextGiveaway.String())
		time.Sleep(time.Until(nextGiveaway.Add(time.Second * 5)))
	}
}
