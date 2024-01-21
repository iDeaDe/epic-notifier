package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

type NotifierConfig struct {
	General struct {
		Channel                   string
		SilentPost                bool
		PostCurrentGamesOnStartup bool
	}
	Timings struct {
		AnnounceRecheckInterval int
		RemindPostDelay         int
		GiveawayRecheckOnFail   int
	}
}

var viperInstance *viper.Viper

func init() {
	viperInstance = viper.New()
}

func createDefaultFile(filepath string, content []byte) error {
	_, err := os.Stat(filepath)
	if err == nil || !os.IsNotExist(err) {
		return err
	}

	file, err := os.Create(filepath)
	defer file.Close()
	if err != nil {
		return err
	}

	_, err = file.Write(content)
	if err != nil {
		return err
	}

	return nil
}

// todo сделать функцию многоразовой
func readConfig(configFilepath string, config interface{}, trackChanges bool) error {
	content, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	err = createDefaultFile(configFilepath, content)
	if err != nil {
		return err
	}

	configDirectory, configFile := filepath.Split(configFilepath)
	nameParts := strings.Split(configFile, ".")

	var fileName, fileExt string

	if len(nameParts) == 2 {
		fileName = nameParts[0]
		fileExt = nameParts[1]
	} else {
		lastElementIndex := len(nameParts) - 1

		fileName = strings.Join(nameParts[:lastElementIndex], ".")
		fileExt = nameParts[lastElementIndex]
	}

	viperInstance.AddConfigPath(configDirectory)
	viperInstance.SetConfigName(fileName)
	viperInstance.SetConfigType(fileExt)

	if err := viperInstance.ReadInConfig(); err != nil {
		return err
	}

	if trackChanges {
		viperInstance.
			OnConfigChange(func(in fsnotify.Event) {
				if in.Has(fsnotify.Write) {
					err := viperInstance.Unmarshal(&config)
					if err != nil {
						Logger().Panic().Err(err).Send()
					}
				}
			})

		viperInstance.WatchConfig()
	}

	err = viperInstance.Unmarshal(&config)
	if err != nil {
		return err
	}

	return nil
}
