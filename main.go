package main

import (
	"flag"
	"github.com/ideade/epic-notifier/epicgames"
	"github.com/ideade/epic-notifier/telegram"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
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

	var config ConfigFile
	config.Name = "config.json"
	config.GetConfig()

	telegramToken := os.Getenv("TELEGRAM_TOKEN")

	if telegramToken == "" {
		log.Panicln("TELEGRAM_TOKEN not found")
	}

	tg := new(telegram.TelegramSettings)
	tg.Token = telegramToken
	tg.ChannelName = config.Content.Channel
	if testChannel != "" {
		tg.ChannelName = testChannel
	}

	tg.RemoveNextPost(strconv.Itoa(config.Content.NextPostId))

	for {
		ga := new(epicgames.Giveaway)
		ga = epicgames.GetGiveaway()
		nextGiveaway := ga.Next

		if nextGiveaway.Before(time.Now()) {
			nextGiveaway = time.Now().Add(time.Hour * 24 * 3)
			postCurrent = false
		}

		if postCurrent {
			for _, game := range ga.CurrentGames {
				log.Printf("Game: %s\n", game.Title)
				tg.Post(&game, silent)
			}
		} else {
			postCurrent = true
			log.Println("Nothing to post")
		}

		log.Println("Creating post about next giveaway")
		config.Content.NextPostId = tg.PostNext(ga)
		_ = config.SaveConfig()
		log.Printf("Next giveaway post ID: %d\n", config.Content.NextPostId)

		ga = nil
		runtime.GC()

		for {
			nextPostInt := strconv.Itoa(config.Content.NextPostId)
			if time.Until(nextGiveaway.Add(time.Second*5)).Hours() < 2 {
				time.Sleep(time.Until(nextGiveaway.Add(time.Second * 5)))
				tg.RemoveNextPost(nextPostInt)
				break
			} else {
				time.Sleep(time.Hour)
			}

			ga = epicgames.GetGiveaway()
			tg.UpdateNext(nextPostInt, ga)
			ga = nil
		}
	}
}
