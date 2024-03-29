package main

import (
	"flag"
	"github.com/ideade/epic-notifier/epicgames"
	"github.com/ideade/epic-notifier/telegram"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

func main() {
	var postCurrent bool
	var silent bool
	var recreateNext bool
	var resendRemind bool
	var testChannel string

	flag.BoolVar(&postCurrent, "c", true, "Specify to not post current games.")
	flag.BoolVar(&silent, "s", false, "Specify to post games silently.")
	flag.BoolVar(&recreateNext, "n", false, "Create new post with games of the next giveaway.")
	flag.BoolVar(&resendRemind, "remind", false, "Resend remind post to the channel.")
	flag.StringVar(&testChannel, "test", "", "Post to the test channel.")
	flag.Parse()

	workDir := os.Getenv("WORKDIR")
	if workDir == "" {
		var err error
		executable, err := os.Executable()
		if err != nil {
			workDir = "."
		} else {
			workDir = filepath.Dir(executable)
		}
	}
	err := os.Chdir(workDir)
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

	tg := new(telegram.Settings)
	tg.Token = telegramToken
	tg.ChannelName = config.Content.Channel
	if testChannel != "" {
		tg.ChannelName = testChannel
	}

	/*
		Пересоздание поста с напоминанием о последнем дне раздачи.
	*/
	if resendRemind {
		if config.Content.RemindPostId == -1 {
			log.Println("Remind post does not exist")
		} else {
			err = tg.RemoveRemind(strconv.Itoa(config.Content.RemindPostId))
			if err != nil {
				log.Println(err)
			}

			config.Content.RemindPostId = -1
			_ = config.SaveConfig()
		}
	}

	/*
		Пересоздание поста с анонсом следующей раздачи.
		Это не работает. В Телеграме нельзя ботом удалять посты старше 48 часов.
	*/
	if recreateNext {
		if config.Content.NextPostId == -1 {
			log.Println("Next giveaway post does not exist")
		} else {
			err = tg.RemoveNextPost(strconv.Itoa(config.Content.NextPostId))
			if err != nil {
				log.Println(err)
			}

			config.Content.NextPostId = -1
			_ = config.SaveConfig()
		}
	}

	for {
		ga := new(epicgames.Giveaway)
		ga = epicgames.GetGiveaway()
		nextGiveaway := ga.Next

		/** Всё, что относится к текущей раздаче **/

		if config.Content.RemindPostId == -1 {
			resendRemind = true
		}

		if nextGiveaway.Before(time.Now()) {
			nextGiveaway = time.Now().Add(time.Hour * 24 * 5)
			postCurrent = true
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

		/** Всё, что относится к следующей раздаче **/

		log.Printf("Next giveaway time: %s\n", nextGiveaway.String())

		if recreateNext || config.Content.NextPostId == -1 {
			log.Println("Creating post about next giveaway")
			config.Content.NextPostId = tg.PostNext(ga)

			_ = config.SaveConfig()
		}
		log.Printf("Next giveaway post ID: %d\n", config.Content.NextPostId)
		runtime.GC()

		for {
			nextPostId := strconv.Itoa(config.Content.NextPostId)
			remindPostId := strconv.Itoa(config.Content.RemindPostId)

			timeUntilNextGiveaway := time.Until(nextGiveaway).Hours()

			if resendRemind && timeUntilNextGiveaway < 6 {
				config.Content.RemindPostId = tg.Remind(ga.CurrentGames)
				resendRemind = false

				_ = config.SaveConfig()
			}

			if time.Until(nextGiveaway.Add(time.Second*5)).Hours() < 2 {
				time.Sleep(time.Until(nextGiveaway.Add(time.Second * 5)))

				err = tg.RemoveNextPost(nextPostId)
				if err != nil {
					log.Println(err)
				} else {
					config.Content.NextPostId = -1
				}

				err = tg.RemoveRemind(remindPostId)
				if err != nil {
					log.Println(err)
				} else {
					config.Content.RemindPostId = -1
				}

				recreateNext = true
				resendRemind = true

				_ = config.SaveConfig()
				break
			} else {
				time.Sleep(time.Hour)
			}

			ga = epicgames.GetGiveaway()
			tg.UpdateNext(nextPostId, ga)
		}
	}
}
