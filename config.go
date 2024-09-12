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

func getMainConfig(path string, trackChanges bool) (*viper.Viper, error) {
	config := viper.New()
	config.AddConfigPath(filepath.Dir(path))
	nameParts := strings.Split(filepath.Base(path), ".")
	config.SetConfigName(strings.Join(nameParts[:len(nameParts)-1], "."))
	config.SetConfigType(nameParts[len(nameParts)-1])

	err := config.ReadInConfig()
	if err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			config.SetDefault(ConfigGeneralLogOutput, "./app.log")
			config.SetDefault(ConfigGeneralChannel, "")
			config.SetDefault(ConfigGeneralSilentPost, false)
			config.SetDefault(ConfigGeneralPostCurrentGamesOnStartup, false)
			config.SetDefault(ConfigGeneralTimezone, "Europe/Moscow")
			config.SetDefault(ConfigGeneralNotificationsChatId, "")

			config.SetDefault(ConfigTimingsAnnounceRecheckInterval, 3600)
			config.SetDefault(ConfigTimingsGiveawayPostDelay, 10)

			config.SetDefault(ConfigEgsApiRecheckOnFail, true)
			config.SetDefault(ConfigEgsApiRecheckOnFailDelay, 60)

			config.SetDefault(ConfigRemindPostEnabled, true)
			config.SetDefault(ConfigRemindPostDelay, 3600*6)

			if err := config.SafeWriteConfig(); err != nil {
				return nil, err
			}

			if err = config.ReadInConfig(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if trackChanges {
		config.WatchConfig()
	}

	return config, err
}
