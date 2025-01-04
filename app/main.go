package main

import (
	"errors"
	"github.com/ideade/epic-notifier/app/currency"
	"github.com/ideade/epic-notifier/app/epicgames"
	"github.com/ideade/epic-notifier/app/logging"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ideade/epic-notifier/app/telegram"
	"github.com/spf13/viper"
)

func main() {
	workdir := getWorkdir()

	dotEnvExists := true
	if _, err := os.Stat(filepath.Join(workdir, ".env")); errors.Is(err, os.ErrNotExist) {
		dotEnvExists = false
	}

	if dotEnvExists && os.Getenv("ENV") != "production" {
		if err := godotenv.Load(filepath.Join(workdir, ".env")); err != nil {
			log.Fatalln(err)
		}
	}

	appEnvironment := logging.EnvDevelopment
	if os.Getenv("ENV") == "production" {
		appEnvironment = logging.EnvProduction
	}

	logger := logging.NewWithEnv(appEnvironment)

	mainConfig, err := getMainConfig(filepath.Join(workdir, "config.toml"), true)
	if err != nil {
		logger.Panic(err.Error())
	}

	if mainConfig.General.Channel == "" {
		logger.Info("Fill config fields")
		os.Exit(0)
	}

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		logger.Error("TELEGRAM_TOKEN env variable is required")
		os.Exit(1)
	}

	_, err = os.Stat(filepath.Join(workdir, ".runtime.json"))
	if errors.Is(err, os.ErrNotExist) {
		file, err := os.OpenFile(filepath.Join(workdir, ".runtime.json"), os.O_WRONLY|os.O_CREATE, 0666)
		if err == nil {
			file.Close()
		}
	}

	runtimeData := viper.New()
	runtimeData.AddConfigPath(workdir)
	runtimeData.SetConfigName(".runtime")
	runtimeData.SetConfigType("json")
	err = runtimeData.ReadInConfig()
	if err != nil {
		logger.Error(err.Error())
	}

	go func() {
		for {
			time.Sleep(time.Second * 5)
			runtimeConfigSaveError := runtimeData.WriteConfig()
			if runtimeConfigSaveError != nil {
				logger.Error(runtimeConfigSaveError.Error())
			}
		}
	}()

	httpClient := &http.Client{Transport: &LoggingRoundTripper{logger, http.DefaultTransport}}

	SetGlobalHttpClient(httpClient)

	epicgames.SetLogger(logger)
	epicgames.SetClient(httpClient)
	epicgames.SetChromeHost(os.Getenv("CHROME_HOST"))

	telegramClient := telegram.NewClient(telegramToken, http.DefaultClient) // using default client because of huge file content in logs
	poster := NewPoster(telegramClient, mainConfig.General.Channel)
	poster.SetSilentMode(mainConfig.General.SilentPost)
	poster.SetTemplateDir(filepath.Join(workdir, "template"))
	if err = poster.SetTimezone(mainConfig.General.Timezone); err != nil {
		logger.Panic(err.Error())
	}

	if mainConfig.General.NotificationsChatId != "" {
		newLogger := logging.AddNotificationHook(logger, telegramClient, mainConfig.General.NotificationsChatId)
		*logger = *newLogger
	}

	updateCurrencies := make(chan bool, 1)

	currenciesToken := os.Getenv("CURRENCIES_TOKEN")
	if currenciesToken != "" {
		currencyUpdater := currency.NewUpdater(httpClient, currenciesToken)
		currencyUpdater.AddPair(*currency.NewPair("RUB", "USD"))
		currencyUpdater.AddPair(*currency.NewPair("KZT", "USD"))
		currencyUpdater.AddPair(*currency.NewPair("KZT", "RUB"))
		if err = currencyUpdater.Update(); err != nil {
			logger.Error(err.Error())
		}

		go func() {
			for {
				if !<-updateCurrencies {
					continue
				}

				err := currencyUpdater.Update()
				if err != nil {
					logger.Error(err.Error())
				}
			}
		}()
	}

	postCurrent := mainConfig.General.PostCurrentGamesOnStartup
	postAnnounce := postCurrent
	removeRemindPost := false

	for {
		remindPostId := runtimeData.GetString("remind_post_id")

		if removeRemindPost && remindPostId != "" {
			_, err := telegramClient.DeleteMessage(&telegram.DeleteMessageRequest{
				ChatId:    mainConfig.General.Channel,
				MessageId: remindPostId,
			})

			if err != nil {
				logger.Error(err.Error())
			} else {
				runtimeData.Set("remind_post_id", "")
				remindPostId = ""
			}

			removeRemindPost = false
		}

		giveaway, err := epicgames.GetGiveaway(mainConfig.Egs.Locale, mainConfig.Egs.Country)
		if err != nil {
			if mainConfig.Egs.RecheckOnFail {
				logger.Error(err.Error())

				time.Sleep(mainConfig.Egs.RecheckOnFailDelay * time.Second)
				continue
			} else {
				logger.Panic(err.Error())
			}
		}

		nextGiveawayDate := giveaway.Next

		if nextGiveawayDate.Before(time.Now()) {
			// todo: подвешивать бота до получения определённой команды
			logger.Panic("incorrect next giveaway date")
		}

		if postCurrent {
			_, err := poster.PostCurrentGames(giveaway.CurrentGames)
			if err != nil {
				logger.Panic(err.Error())
				continue
			}
		}

		if postAnnounce {
			_, err := poster.PostAnnounce(giveaway)
			if err != nil {
				logger.Panic(err.Error())
			}
		}

		if mainConfig.RemindPost.Enabled && remindPostId == "" {
			sleepTime := time.Until(nextGiveawayDate) - mainConfig.RemindPost.Delay*time.Second
			sleep(logger, sleepTime*time.Second)

			newRemindPostId, err := poster.PostRemind(giveaway)
			if err != nil {
				logger.Panic(err.Error())
			} else {
				remindPostId = newRemindPostId
				runtimeData.Set("remind_post_id", remindPostId)
			}
		} else {
			// 30 seconds to update currency rates
			sleepTime := time.Until(nextGiveawayDate) - time.Second*30
			sleep(logger, sleepTime)
		}

		updateCurrencies <- true
		removeRemindPost = true
		postCurrent = true
		postAnnounce = true

		sleepTime := time.Until(nextGiveawayDate) + time.Second*mainConfig.Timings.GiveawayPostDelay
		sleep(logger, sleepTime)
	}
}
