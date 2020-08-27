package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Channel string `yaml:"channel"`
}

func GetConfig(filename string) *Config {
	cfgFile, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	if os.IsNotExist(err) {
		cfgFile = createDefaultConfig(filename)
	} else if err != nil {
		log.Panicf("Can't open the config file.\nError: %v", err)
	}
	defer cfgFile.Close()

	var fileSize int64
	fileStat, err := cfgFile.Stat()
	if err != nil {
		fileSize = 4096
	} else {
		fileSize = fileStat.Size()
	}

	fileContent := make([]byte, fileSize)
	_, err = cfgFile.Read(fileContent)

	config := new(Config)
	err = yaml.Unmarshal(fileContent, config)
	// todo: fix this shit
	if err != nil {
		log.Panicf("Error in unmarshal config.\n%v", err)
	}

	return config
}

func createDefaultConfig(filename string) *os.File {
	defaultConfig := Config{
		Channel: "@epicgiveaways",
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Panicf("Can't create the config file.\nError: %v", err)
	}

	content, err := yaml.Marshal(defaultConfig)
	if err != nil {
		log.Panicf("Can't marshal default config.\nError: %v", err)
	}

	_, err = file.Write(content)
	if err != nil {
		log.Panicln("Seems like config file is unwritable")
	}

	return file
}
