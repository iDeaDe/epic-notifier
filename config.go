package main

import (
	"encoding/json"
	"log"
	"os"
)

type ConfigFile struct {
	Name    string
	Content *Config
}

type Config struct {
	Channel    string `json:"channel"`
	NextPostId int    `json:"next_post_id"`
}

var defaultConfig = Config{
	Channel:    "",
	NextPostId: 0,
}

func (file *ConfigFile) SaveConfig() error {
	cfgFile, err := os.OpenFile(file.Name, os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	content, err := json.Marshal(file.Content)
	if err != nil {
		return err
	}
	_, err = cfgFile.Write(content)
	if err != nil {
		return err
	}

	return nil
}

func (file *ConfigFile) GetConfig() {
	cfgFile, err := os.OpenFile(file.Name, os.O_RDONLY, os.ModePerm)

	if os.IsNotExist(err) {
		createDefaultConfig(file.Name)
		defaultCopy := defaultConfig
		file.Content = &defaultCopy
		return
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
	err = json.Unmarshal(fileContent, config)
	if err != nil {
		log.Panicf("Error in unmarshal config.\n%v", err)
	}
	file.Content = config
}

func createDefaultConfig(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Panicf("Can't create the config file.\nError: %v", err)
	}
	defer file.Close()

	content, err := json.Marshal(defaultConfig)
	if err != nil {
		log.Panicf("Can't marshal default config.\nError: %v", err)
	}

	_, err = file.Write(content)
	if err != nil {
		log.Panicln("Seems like config file is unwritable")
	}

	log.Printf("Fill all fields in created config file(%s)\n", filename)
	os.Exit(0)
}
