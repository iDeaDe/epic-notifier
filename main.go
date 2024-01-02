package main

import (
	"errors"
	"github.com/ideade/epic-notifier/app"
	"github.com/spf13/viper"
	"log"
	"time"
)

func main() {
	application, err := app.InitApp()
	if err != nil {
		log.Panicln(err)
	}

	runtimeData := viper.New()
	runtimeData.AddConfigPath(application.WorkDir)
	runtimeData.SetConfigName(".runtime")
	runtimeData.SetConfigType("json")

	if err := runtimeData.ReadInConfig(); err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
		app.Logger().Panic().Err(err).Send()
	}

	for {
		time.Sleep(3 * time.Second)
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

		tg.Send(&req)
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
}
