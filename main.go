package main

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ideade/epic-notifier/epicgames"
	"github.com/ideade/epic-notifier/telegram"
	"github.com/spf13/viper"
)

func main() {
	workdir := getWorkdir()

	mainConfig, err := mainConfig(filepath.Join(workdir, "config.toml"), true)
	if err != nil {
		Logger().Panic().Stack().Err(err).Send()
	}

	if mainConfig.GetString(ConfigGeneralChannel) == "" {
		Logger().Print("Fill config fields")
		os.Exit(0)
	}

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		Logger().Error().Msg("TELEGRAM_TOKEN env variable is required")
		os.Exit(1)
	}

	runtimeData := viper.New()
	runtimeData.AddConfigPath(workdir)
	runtimeData.SetConfigName(".runtime")

	saveConfig := make(chan bool, 1)

	go func() {
		for {
			if !<-saveConfig {
				continue
			}

			err := runtimeData.SafeWriteConfig()
			if err != nil {
				Logger().Panic().Err(err).Send()
			}
		}
	}()

	epicgames.SetLogger(Logger())

	telegramClient := telegram.NewClient(telegramToken, http.DefaultClient)
	poster := NewPoster(telegramClient, mainConfig.GetString(ConfigGeneralChannel))
	if err = poster.SetTimezone(mainConfig.GetString(ConfigGeneralTimezone)); err != nil {
		Logger().Panic().Err(err).Send()
	}

	if mainConfig.GetString(ConfigGeneralNotificationsChatId) != "" {
		AddLoggerHook(NewNotifyHook(telegramClient, mainConfig.GetString(ConfigGeneralNotificationsChatId)))
	}

	postCurrent := mainConfig.GetBool(ConfigGeneralPostCurrentGamesOnStartup)
	postAnnounce := postCurrent
	removeRemindPost := false

	for {
		remindPostId := runtimeData.GetString("remind_post_id")
		if removeRemindPost && remindPostId != "" {
			_, err := telegramClient.DeleteMessage(&telegram.DeleteMessageRequest{
				ChatId:    mainConfig.GetString(ConfigGeneralChannel),
				MessageId: remindPostId,
			})

			if err != nil {
				Logger().Err(err).Send()
			} else {
				runtimeData.Set("remind_post_id", nil)
				saveConfig <- true
			}
		}

		giveaway, err := epicgames.GetGiveaway()
		if err != nil {
			if mainConfig.GetBool(ConfigEgsApiRecheckOnFail) {
				Logger().Err(err).Send()

				time.Sleep(mainConfig.GetDuration(ConfigEgsApiRecheckOnFailDelay))
				continue
			} else {
				Logger().Panic().Err(err).Send()
			}
		}

		if postCurrent {
			_, err := poster.PostCurrentGames(giveaway.CurrentGames)
			if err != nil {
				Logger().Panic().Err(err).Send()
				continue
			}

			postCurrent = false
		}

		if postAnnounce {
			_, err := poster.PostAnnounce(giveaway)
			if err != nil {
				Logger().Panic().Err(err).Send()
			}
		}

		nextGiveawayDate := giveaway.Next

		if nextGiveawayDate.Before(time.Now()) {
			// todo: подвешивать бота до получения определённой команды
			Logger().Panic().Err(errors.New("incorrect next giveaway date")).Send()
		}

		if mainConfig.GetBool(ConfigRemindPostEnabled) {
			sleepTime := time.Until(nextGiveawayDate).Seconds() - mainConfig.GetFloat64(ConfigRemindPostDelay)
			time.Sleep(time.Second * time.Duration(sleepTime))

			newRemindPostId, err := poster.PostRemind()
			if err != nil {
				Logger().Panic().Err(err).Send()
			} else {
				remindPostId = newRemindPostId
				runtimeData.Set("remind_post_id", remindPostId)
				saveConfig <- true
			}
		}

		removeRemindPost = true
		postAnnounce = true

		time.Sleep(time.Until(nextGiveawayDate) + mainConfig.GetDuration(ConfigTimingsGiveawayPostDelay))
	}
}

/*logOut := os.Stderr
logFilePath := filepath.Join(filepath.Dir(application.WorkDir), "app.log")

fmt.Printf("Log file location %s", logFilePath)

logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
if err == nil {
	logOut = logFile
	log.SetOutput(logFile)
}
defer logFile.Close()

logger = log.New(logOut, "[Main] ", log.LstdFlags|log.Lshortfile)
if err != nil {
	logger.Fatalln(err)
}

telegram.SetLogger(log.New(logOut, "[Telegram] ", log.LstdFlags|log.Lshortfile))
epicgames.SetLogger(log.New(logOut, "[Epicgames] ", log.LstdFlags|log.Lshortfile))

err = os.Chdir(application.WorkDir)
if err != nil {
	logger.Panicln(err)
}

config := viper.New()
config.SetConfigName("config")
config.SetConfigType("toml")
config.AddConfigPath(application.WorkDir)
err = config.ReadInConfig()
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

/*
	Пересоздание поста с напоминанием о последнем дне раздачи.

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

tpl, err := template.New("game.gohtml").
	Funcs(
		template.FuncMap{
			"month": epicgames.GetMonth,
		},
	).
	ParseFiles(filepath.Join("template", "game.gohtml"))

ga, err := epicgames.GetExtendedGiveaway()
if err != nil {
	log.Panicln(err)
}

for _, game := range ga.NextGames {
	buf := bytes.NewBufferString("")
	tpl.Execute(buf, game)

	queryParams := map[string]string{
		"chat_id":    tg.ChannelName,
		"photo":      game.Image,
		"parse_mode": "HTML",
		"caption":    buf.String(),
	}

	req := telegram.Request{
		Method: telegram.MethodGet,
		Name:   "sendPhoto",
		Params: &queryParams,
		Body:   nil,
	}

	tg.send(&req)
}

ga := new(epicgames.Giveaway)

for {
	giveaway, err := epicgames.GetExtendedGiveaway()
	if err != nil {
		logger.Println(err)

		time.Sleep(time.Minute)
		continue
	}

}*/
