package app

import (
	"fmt"
	"github.com/ideade/epic-notifier/epicgames"
	"github.com/ideade/epic-notifier/telegram"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"
)

var args BotArgs
var configFileName = "config.json"
var telegramToken string

func init() {
	args = GetArgs()

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

	telegramToken = os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		log.Panicln("TELEGRAM_TOKEN not found")
	}
}

func Start() {
	var config ConfigFile
	config.Path = configFileName

	// Получаем всю необходимую информацию
	config.GetConfig()

	// Заполняем всю информацию, необходимую для работы бота
	tg := new(telegram.TelegramSettings)
	tg.Token = telegramToken
	tg.ChannelName = config.Content.Channel
	if args.TestChannelName != "" {
		tg.ChannelName = args.TestChannelName
	}

	if args.RecreateNext {
		err := tg.RemoveNextPost(strconv.Itoa(config.Content.NextPostId))
		if err != nil {
			log.Println(err)
		}
	}

	for {
		ga := new(epicgames.Giveaway)
		ga = epicgames.GetGiveaway()
		nextGiveaway := ga.Next

		if nextGiveaway.Before(time.Now()) {
			nextGiveaway = time.Now().Add(time.Hour * 24 * 3)
			args.PostCurrent = false
		}

		if args.PostCurrent {
			for _, game := range ga.CurrentGames {
				log.Printf("Game: %s\n", game.Title)
				tg.Post(&game, args.PostSilently)
			}
		} else {
			args.PostCurrent = true
			log.Println("Nothing to post")
		}

		log.Printf("Next giveaway time: %s\n", nextGiveaway.String())

		if args.RecreateNext {
			log.Println("Creating post about next giveaway")
			config.Content.NextPostId = tg.PostNext(ga)
			_ = config.SaveConfig()
		}
		log.Printf("Next giveaway post ID: %d\n", config.Content.NextPostId)

		ga = nil
		args.RecreateNext = true
		runtime.GC()

		fmt.Println(time.Now())

		for {
			nextPostId := strconv.Itoa(config.Content.NextPostId)
			if time.Until(nextGiveaway.Add(time.Second*5)).Hours() < 2 {
				time.Sleep(time.Until(nextGiveaway.Local().Add(time.Second * 5)))
				err := tg.RemoveNextPost(nextPostId)
				if err != nil {
					log.Println(err)
				}
				break
			} else {
				time.Sleep(time.Hour)
			}

			ga = epicgames.GetGiveaway()
			tg.UpdateNext(nextPostId, ga)
			ga = nil
		}
	}
}
