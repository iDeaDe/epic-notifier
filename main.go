package main

import (
	"github.com/ideade/epic-notifier/epicgames"
	"github.com/ideade/epic-notifier/telegram"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"time"
)

func main() {
	workdir := getWorkdir()
	config := NotifierConfig{}

	err := readConfig(filepath.Join(workdir, "config.toml"), config, true)
	if err != nil {
		Logger().Panic().Stack().Err(err).Send()
	}

	koanf.Provider()

	if config.General.Channel == "" {
		Logger().Print("Fill config fields")
		os.Exit(0)
	}

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		Logger().Panic().Msg("TELEGRAM_TOKEN env variable is required")
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

	telegramClient := telegram.NewClient(telegramToken)

	postCurrent := config.General.PostCurrentGamesOnStartup

	for {
		giveaway, err := epicgames.GetExtendedGiveaway()
		if err != nil {
			Logger().Err(err).Send()
			time.Sleep(time.Duration(config.Timings.GiveawayRecheckOnFail))
			continue
		}

		if postCurrent {
			for _, game := range giveaway.CurrentGames {
				_, err := PostGame(telegramClient, config.General.Channel, &game)
				if err != nil {
					Logger().Err(err)
					continue
				}
			}
		}
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
