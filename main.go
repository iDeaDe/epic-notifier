package main

import (
	"errors"
	"github.com/ideade/epic-notifier/currency"
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
	runtimeData.SetConfigType("json")

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

	httpClient := &http.Client{Transport: &LoggingRoundTripper{http.DefaultTransport}}

	SetGlobalHttpClient(httpClient)

	epicgames.SetLogger(Logger())
	epicgames.SetClient(httpClient)

	telegramClient := telegram.NewClient(telegramToken, httpClient)
	poster := NewPoster(telegramClient, mainConfig.GetString(ConfigGeneralChannel))
	if err = poster.SetTimezone(mainConfig.GetString(ConfigGeneralTimezone)); err != nil {
		Logger().Panic().Err(err).Send()
	}

	if mainConfig.GetString(ConfigGeneralNotificationsChatId) != "" {
		AddLoggerHook(NewNotifyHook(telegramClient, mainConfig.GetString(ConfigGeneralNotificationsChatId)))
	}

	updateCurrencies := make(chan bool, 1)

	currenciesToken := os.Getenv("CURRENCIES_TOKEN")
	if currenciesToken != "" {
		currencyUpdater := currency.NewUpdater(httpClient, currenciesToken)
		currencyUpdater.AddPair(*currency.NewPair("RUB", "USD"))
		currencyUpdater.AddPair(*currency.NewPair("KZT", "USD"))
		currencyUpdater.AddPair(*currency.NewPair("KZT", "RUB"))
		if err = currencyUpdater.Update(); err != nil {
			Logger().Error().Err(err).Send()
		}

		go func() {
			for {
				if !<-updateCurrencies {
					continue
				}

				err := currencyUpdater.Update()
				if err != nil {
					Logger().Error().Err(err).Send()
				}
			}
		}()
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
		} else {
			time.Sleep(time.Until(nextGiveawayDate) - time.Second*30)
		}

		updateCurrencies <- true
		removeRemindPost = true
		postAnnounce = true

		time.Sleep(time.Until(nextGiveawayDate) + mainConfig.GetDuration(ConfigTimingsGiveawayPostDelay))
	}
}
