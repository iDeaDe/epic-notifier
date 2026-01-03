package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ideade/epic-notifier/app/currency"
	"github.com/ideade/epic-notifier/app/currency/exchangerate"
	"github.com/ideade/epic-notifier/app/epicgames"
	"github.com/ideade/epic-notifier/app/logging"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/ideade/epic-notifier/app/telegram"
	"github.com/spf13/viper"
)

func main() {
	workdir := getWorkdir()

	//region .env
	dotEnvExists := true
	if _, err := os.Stat(filepath.Join(workdir, ".env")); errors.Is(err, os.ErrNotExist) {
		dotEnvExists = false
	}

	if dotEnvExists && os.Getenv("ENV") != "production" {
		if err := godotenv.Load(filepath.Join(workdir, ".env")); err != nil {
			log.Fatalln(err)
		}
	}
	//endregion

	//region logger
	appEnvironment := logging.EnvDevelopment
	if os.Getenv("ENV") == "production" {
		appEnvironment = logging.EnvProduction
	}

	logger := logging.NewWithEnv(appEnvironment)
	//endregion

	var err error

	if err = loadCountriesMap(workdir); err != nil {
		logger.Panic("failed to load countries map", zap.Error(err))
	}

	//region main config loading and required data check
	var mainConfig *Config
	if mainConfig, err = getMainConfig(filepath.Join(workdir, "config.toml"), false); err != nil {
		logger.Panic("failed to load main config", zap.Error(err))
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
	//endregion

	//region runtime data file
	_, err = os.Stat(filepath.Join(workdir, ".runtime.json"))
	if errors.Is(err, os.ErrNotExist) {
		var file *os.File
		file, err = os.OpenFile(filepath.Join(workdir, ".runtime.json"), os.O_WRONLY|os.O_CREATE, 0666)
		if err == nil {
			_ = file.Close()
		}
	}

	runtimeData := viper.New()
	runtimeData.AddConfigPath(workdir)
	runtimeData.SetConfigName(".runtime")
	runtimeData.SetConfigType("json")
	err = runtimeData.ReadInConfig()
	if err != nil {
		logger.Error("failed to read runtime data", zap.Error(err))
	}

	go func() {
		for {
			time.Sleep(time.Second * 5)
			runtimeConfigSaveError := runtimeData.WriteConfig()
			if runtimeConfigSaveError != nil {
				logger.Error("failed to write new runtime data", zap.Error(runtimeConfigSaveError))
			}
		}
	}()
	//endregion

	httpClient := &http.Client{Transport: &LoggingRoundTripper{logger, http.DefaultTransport}}

	SetGlobalHttpClient(httpClient)

	epicgames.SetLogger(logger)
	epicgames.SetClient(httpClient)

	telegramClient := telegram.NewClient(telegramToken, http.DefaultClient) // using default client because of huge file content in logs

	//region telegram channel games poster
	poster := NewPoster(telegramClient, mainConfig.General.Channel)
	poster.SetSilentMode(mainConfig.General.SilentPost)
	poster.SetTemplateDir(filepath.Join(workdir, "template"))

	var timezone *time.Location
	if timezone, err = time.LoadLocation(mainConfig.General.Timezone); err == nil {
		poster.SetTimezone(timezone)
	} else {
		logger.Panic("failed to load timezone", zap.Error(err))
	}

	poster.SetTimezone(timezone)
	//endregion

	if mainConfig.General.NotificationsChatId != "" {
		newLogger := logging.AddNotificationHook(logger, telegramClient, mainConfig.General.NotificationsChatId)
		*logger = *newLogger
	}

	//region currencies updater
	currenciesToken := os.Getenv("CURRENCIES_TOKEN")

	if currenciesToken != "" {
		currencyUpdater = exchangerate.NewBackgroundUpdater(httpClient, currenciesToken)
		//currencyUpdater = currencyapi.NewBackgroundUpdater(httpClient, currenciesToken)
	} else {
		currencyUpdater = &NopCurrencyUpdater{}
	}

	currencyUpdater.SetErrorHandler(func(err error) {
		logger.Error("failed to update currencies", zap.Error(err))
	})
	currencyUpdater.AddPair(*currency.NewPair("KZT", "USD"))
	currencyUpdater.AddPair(*currency.NewPair("KZT", "RUB"))
	currencyUpdater.Update()
	//endregion

	postCurrent := mainConfig.General.PostCurrentGamesOnStartup
	postAnnounce := postCurrent
	removeRemindPost := false

	for {
		remindPostId := runtimeData.GetString("remind_post_id")

		if removeRemindPost && remindPostId != "" {
			_, err = telegramClient.DeleteMessage(&telegram.DeleteMessageRequest{
				ChatId:    mainConfig.General.Channel,
				MessageId: remindPostId,
			})

			if err != nil {
				logger.Error("failed to remove telegram post", zap.Error(err))
			} else {
				runtimeData.Set("remind_post_id", "")
				remindPostId = ""
			}

			removeRemindPost = false
		}

		var giveaway *epicgames.Giveaway
		giveaway, err = epicgames.GetGiveaway(mainConfig.Egs.Locale, mainConfig.Egs.Country)
		if err != nil {
			if mainConfig.Egs.RecheckOnFail {
				logger.Error("failed to get egs giveaway", zap.Error(err))

				time.Sleep(mainConfig.Egs.RecheckOnFailDelay * time.Second)
				continue
			} else {
				logger.Panic("failed to get egs giveaway", zap.Error(err))
			}
		}

		nextGiveawayDate := giveaway.Next

		if nextGiveawayDate.Before(time.Now()) {
			// todo: подвешивать бота до получения определённой команды
			logger.Panic("incorrect next giveaway date")
		}

		if postCurrent {
			_, err = poster.PostCurrentGames(giveaway.CurrentGames)
			if err != nil {
				logger.Panic("failed to post game", zap.Error(err))
				continue
			}
		}

		if postAnnounce {
			_, err = poster.PostAnnounce(giveaway)
			if err != nil {
				logger.Panic("failed to post announce", zap.Error(err))
			}
		}

		if mainConfig.RemindPost.Enabled && remindPostId == "" {
			sleepTime := time.Until(nextGiveawayDate) - mainConfig.RemindPost.Delay*time.Second
			sleep(logger, sleepTime)

			var newRemindPostId string
			newRemindPostId, err = poster.PostRemind(giveaway)
			if err != nil {
				logger.Panic("failed to post remind post", zap.Error(err))
			} else {
				remindPostId = newRemindPostId
				runtimeData.Set("remind_post_id", remindPostId)
			}
		}

		// 30 seconds to update currency rates
		sleepTime := time.Until(nextGiveawayDate) - time.Second*30
		sleep(logger, sleepTime)

		currencyUpdater.Update()

		removeRemindPost = true
		postCurrent = true
		postAnnounce = true

		sleepTime = time.Until(nextGiveawayDate) + time.Second*mainConfig.Timings.GiveawayPostDelay
		sleep(logger, sleepTime)
	}
}
