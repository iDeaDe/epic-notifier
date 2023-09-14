package main

import (
	"flag"
	"fmt"
	"github.com/ideade/epic-notifier/epicgames"
	"github.com/ideade/epic-notifier/telegram"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

var logger *log.Logger

type NotifierConfig struct {
	Channel             string `json:"channel"`
	NotificationsUserId int64  `json:"notifications_user_id"`
	NextPostId          int    `json:"next_post_id"`
	RemindPostId        int    `json:"remind_post_id"`
}

func (c *NotifierConfig) GetFilePath() string {
	return "config.json"
}

func getDefaultConfig() *NotifierConfig {
	return &NotifierConfig{
		Channel:             "",
		NotificationsUserId: -1,
		NextPostId:          -1,
		RemindPostId:        -1,
	}
}

func createPidFile() error {
	var err error

	appTempDir := filepath.Join(os.TempDir(), "epic-notifier")

	if _, err := os.Stat(appTempDir); err != nil && os.IsNotExist(err) {
		err = os.Mkdir(appTempDir, 0770)
	}
	if err != nil {
		return err
	}

	pidFile, err := os.Create(filepath.Join(appTempDir, ".running.pid"))
	if err != nil {
		return err
	}

	_, err = pidFile.WriteString(strconv.Itoa(os.Getpid()))

	return err
}

func main() {
	err := createPidFile()
	if err != nil {
		logger.Fatalln(err)
	}

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

	logOut := os.Stderr

	logDir := filepath.Dir(workDir)
	logFile, err := os.OpenFile(filepath.Join(logDir, "app.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		logOut = logFile
		log.SetOutput(logFile)
	}

	logger = log.New(logOut, "[Main] ", log.LstdFlags|log.Lshortfile)
	if err != nil {
		logger.Fatalln(err)
	}

	telegram.SetLogger(log.New(logOut, "[Telegram] ", log.LstdFlags|log.Lshortfile))
	epicgames.SetLogger(log.New(logOut, "[Epicgames] ", log.LstdFlags|log.Lshortfile))

	err = os.Chdir(workDir)
	if err != nil {
		logger.Panicln(err)
	}

	cfg := new(NotifierConfig)
	err = ReadConfig(cfg)
	if os.IsNotExist(err) {
		err = SaveConfig(getDefaultConfig())
		fmt.Println("Fill config field and rerun the bot")
		os.Exit(0)
	}
	if err != nil {
		logger.Fatalln(err)
	}

	go func() {
		err := logFile.Sync()
		if err != nil {
			log.Fatalln(err)
		}

		time.Sleep(time.Second * 3)
	}()

	autosaver := NewAutosaver(cfg, time.Second*5)
	autosaver.Start()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)

		<-sig

		err := autosaver.Stop()
		if err != nil {
			logger.Println(err)
		}

		for autosaver.Running {
			time.Sleep(time.Second)
		}

		err = logFile.Close()
		if err != nil {
			log.Fatalln(err)
		}

		os.Exit(0)
	}()

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		logger.Panicln("TELEGRAM_TOKEN not found")
	}

	tg := new(telegram.Telegram)
	tg.Token = telegramToken
	tg.ChannelName = cfg.Channel
	if testChannel != "" {
		tg.ChannelName = testChannel
	}

	/*
		Пересоздание поста с напоминанием о последнем дне раздачи.
	*/
	if resendRemind {
		if cfg.RemindPostId == -1 {
			logger.Println("Remind post does not exist")
		} else {
			err = tg.RemoveRemind(cfg.RemindPostId)
			if err != nil {
				logger.Println(err)
			}

			cfg.RemindPostId = -1
		}
	} else if cfg.RemindPostId == -1 {
		resendRemind = true
	}

	/*
		Пересоздание поста с анонсом следующей раздачи.
		Это не работает. В Телеграме нельзя ботом удалять посты старше 48 часов.
	*/
	if recreateNext {
		if cfg.NextPostId == -1 {
			logger.Println("Next giveaway post does not exist")
		} else {
			err = tg.RemoveNextPost(cfg.NextPostId)
			if err != nil {
				logger.Println(err)
			}

			cfg.NextPostId = -1
		}
	} else if cfg.NextPostId == -1 {
		recreateNext = true
	}

	for {
		ga := new(epicgames.Giveaway)
		ga, err = epicgames.GetGiveaway()
		if err != nil {
			logger.Panicln(err)
		}

		nextGiveaway := ga.Next

		/** Всё, что относится к текущей раздаче **/

		if nextGiveaway.Before(time.Now()) {
			nextGiveaway = time.Now().Add(time.Hour * 24 * 5)
			postCurrent = true
			logger.Println(fmt.Sprintf("Next giveaway time was replaced with %s", nextGiveaway.String()))
		}

		if postCurrent {
			for _, game := range ga.CurrentGames {
				logger.Println(fmt.Sprintf("Game: %s", game.Title))
				err = tg.Post(&game, silent)
				if err != nil {
					logger.Println(err)
				}
			}
		} else {
			postCurrent = true
			logger.Println("Nothing to post")
		}

		/** Всё, что относится к следующей раздаче **/

		logger.Println(fmt.Sprintf("Next giveaway time: %s", nextGiveaway.String()))

		if recreateNext {
			logger.Println("Creating post about next giveaway")
			cfg.NextPostId, err = tg.PostNext(ga)
			if err != nil {
				logger.Println(err)
			}

			if err = SaveConfig(cfg); err != nil {
				logger.Fatalln(err)
			}
		}

		logger.Printf("Next giveaway post ID: %d\n", cfg.NextPostId)
		runtime.GC()

		for {
			timeUntilNextGiveaway := time.Until(nextGiveaway).Hours()

			if resendRemind && timeUntilNextGiveaway < 6 {
				cfg.RemindPostId, err = tg.Remind(ga.CurrentGames)
				if err != nil {
					logger.Println(err)
				}
				resendRemind = false

				if err = SaveConfig(cfg); err != nil {
					logger.Fatalln(err)
				}
			}

			if time.Until(nextGiveaway.Add(time.Second*5)).Hours() >= 2 {
				time.Sleep(time.Hour)
				break
			}

			if time.Until(nextGiveaway.Add(time.Second*5)).Hours() < 2 {
				time.Sleep(time.Until(nextGiveaway.Add(time.Second * 5)))

				err = tg.RemoveNextPost(cfg.NextPostId)
				if err != nil {
					logger.Println(err)
				} else {
					cfg.NextPostId = -1
				}

				err = tg.RemoveRemind(cfg.RemindPostId)
				if err != nil {
					logger.Println(err)
				} else {
					cfg.RemindPostId = -1
				}

				recreateNext = true
				resendRemind = true

				if err = SaveConfig(cfg); err != nil {
					logger.Fatalln(err)
				}
				break
			} else {
				time.Sleep(time.Hour)
			}

			ga, err = epicgames.GetGiveaway()
			if err != nil {
				logger.Panicln(err)
			}

			err = tg.UpdateNext(cfg.NextPostId, ga)
			if err != nil {
				logger.Println(err)
			}
		}
	}
}
