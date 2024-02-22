package main

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	ConfigGeneralLogOutput                 = "general.log_output"
	ConfigGeneralChannel                   = "general.channel"
	ConfigGeneralSilentPost                = "general.silent_post"
	ConfigGeneralPostCurrentGamesOnStartup = "general.post_current_games_on_startup"
	ConfigGeneralTimezone                  = "general.timezone"
	ConfigGeneralNotificationsChatId       = "general.notifications_chat_id"

	ConfigTimingsAnnounceRecheckInterval = "timings.announce_recheck_interval"
	ConfigTimingsGiveawayPostDelay       = "timings.giveaway_post_delay"

	ConfigEgsApiRecheckOnFail      = "egs_api.recheck_on_fail"
	ConfigEgsApiRecheckOnFailDelay = "egs_api.recheck_on_fail_delay"

	ConfigRemindPostEnabled = "remind_post.enabled"
	ConfigRemindPostDelay   = "remind_post.delay"
)

func mainConfig(path string, trackChanges bool) (*viper.Viper, error) {
	mainConfig := viper.New()
	mainConfig.AddConfigPath(filepath.Dir(path))
	nameParts := strings.Split(filepath.Base(path), ".")
	mainConfig.SetConfigName(strings.Join(nameParts[:len(nameParts)-1], "."))
	mainConfig.SetConfigType(nameParts[len(nameParts)-1])

	err := mainConfig.ReadInConfig()
	if err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			mainConfig.SetDefault(ConfigGeneralLogOutput, "./app.log")
			mainConfig.SetDefault(ConfigGeneralChannel, "")
			mainConfig.SetDefault(ConfigGeneralSilentPost, false)
			mainConfig.SetDefault(ConfigGeneralPostCurrentGamesOnStartup, false)
			mainConfig.SetDefault(ConfigGeneralTimezone, "Europe/Moscow")
			mainConfig.SetDefault(ConfigGeneralNotificationsChatId, "")

			mainConfig.SetDefault(ConfigTimingsAnnounceRecheckInterval, 3600)
			mainConfig.SetDefault(ConfigTimingsGiveawayPostDelay, 10)

			mainConfig.SetDefault(ConfigEgsApiRecheckOnFail, true)
			mainConfig.SetDefault(ConfigEgsApiRecheckOnFailDelay, 60)

			mainConfig.SetDefault(ConfigRemindPostEnabled, true)
			mainConfig.SetDefault(ConfigRemindPostDelay, 3600*6)

			if err := mainConfig.SafeWriteConfig(); err != nil {
				return nil, err
			}

			if err = mainConfig.ReadInConfig(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if trackChanges {
		mainConfig.WatchConfig()
	}

	return mainConfig, err
}
