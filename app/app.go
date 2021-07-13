package app

import (
	"github.com/ideade/epic-notifier/epicgames"
	"github.com/ideade/epic-notifier/telegram"
	"github.com/ideade/epic-notifier/utils"
	"log"
	"runtime"
	"strconv"
	"time"
)

type Parameters struct {
	PostCurrent bool
	Silent      bool
	Next        bool
	Remind      bool
	TestChannel string
	ConfigFile  string
}

type App struct {
	remindSent bool
	config     ConfigFile
	params     Parameters
	tg         *telegram.TelegramSettings
}

var logger = utils.NewLogger("[Main]")

func init() {
	/**
	Проверка соединения с сервисами
	*/
	logger.Println("EGS Giveaways bot is starting")
	logger.Println("Connecting to EGS")
	err := epicgames.CheckConnection()
	if err == nil {
		logger.Println("Connection with EGS is OK")
	} else {
		logger.Panicln(err)
	}

	logger.Println("Connecting to Telegram API")
	err = telegram.CheckConnection()
	if err == nil {
		logger.Println("Connection with Telegram API is OK")
	} else {
		logger.Panicln(err)
	}
}

func (app *App) Loop() {
	for {
		ga := new(epicgames.Giveaway)
		ga = epicgames.GetGiveaway()
		nextGiveaway := ga.Next

		/** Всё, что относится к текущей раздаче **/

		if nextGiveaway.Before(time.Now()) {
			nextGiveaway = time.Now().Add(time.Hour * 24 * 3)
			app.params.PostCurrent = false
		}

		if app.params.PostCurrent {
			for _, game := range ga.CurrentGames {
				log.Printf("Game: %s\n", game.Title)
				app.tg.Post(&game, app.params.Silent)
			}
		} else {
			app.params.PostCurrent = true
			log.Println("Nothing to post")
		}

		/** Всё, что относится к следующей раздаче **/

		log.Printf("Next giveaway time: %s\n", nextGiveaway.String())

		if app.params.Next {
			log.Println("Creating post about next giveaway")
			app.config.Content.NextPostId = app.tg.PostNext(ga)

			_ = app.config.SaveConfig()
		}
		log.Printf("Next giveaway post ID: %d\n", app.config.Content.NextPostId)

		app.params.Next = true
		app.remindSent = false
		runtime.GC()

		for {
			nextPostId := strconv.FormatUint(app.config.Content.NextPostId, 10)
			remindPostId := strconv.FormatUint(app.config.Content.RemindPostId, 10)

			if !app.remindSent && time.Until(nextGiveaway).Hours() < 6 {
				app.config.Content.RemindPostId = app.tg.Remind(ga.CurrentGames)
				app.remindSent = true

				_ = app.config.SaveConfig()
			}

			if time.Until(nextGiveaway.Add(time.Second*5)).Hours() < 2 {
				time.Sleep(time.Until(nextGiveaway.Add(time.Second * 5)))

				err := app.tg.RemoveNextPost(nextPostId)
				if err != nil {
					log.Println(err)
				} else {
					app.config.Content.NextPostId = 0
				}

				err = app.tg.RemoveRemind(remindPostId)
				if err != nil {
					log.Println(err)
				} else {
					app.config.Content.RemindPostId = 0
				}

				_ = app.config.SaveConfig()
				break
			} else {
				time.Sleep(time.Hour)
			}

			ga = epicgames.GetGiveaway()
			app.tg.UpdateNext(nextPostId, ga)
		}
	}
}
