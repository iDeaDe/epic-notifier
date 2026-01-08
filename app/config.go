package main

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	General struct {
		LogOutput                 string
		Channel                   string
		SilentPost                bool
		PostCurrentGamesOnStartup bool
		Timezone                  string
		NotificationsChatId       string
	}
	Timings struct {
		AnnounceRecheckInterval time.Duration
		GiveawayPostDelay       time.Duration
	}
	Egs struct {
		RecheckOnFail      bool
		RecheckOnFailDelay time.Duration
		Locale             string
		Country            string
	}
	RemindPost struct {
		Enabled bool
		Delay   time.Duration
	}
}

func getMainConfig(path string, trackChanges bool) (*Config, error) {
	config := viper.New()
	config.AddConfigPath(filepath.Dir(path))
	nameParts := strings.Split(filepath.Base(path), ".")
	config.SetConfigName(strings.Join(nameParts[:len(nameParts)-1], "."))
	config.SetConfigType(nameParts[len(nameParts)-1])

	err := config.ReadInConfig()
	if err != nil {
		return nil, err
	}

	targetStruct := &Config{}

	err = config.Unmarshal(targetStruct)
	if err != nil {
		return nil, err
	}

	if trackChanges {
		config.WatchConfig()
	}

	return targetStruct, err
}
