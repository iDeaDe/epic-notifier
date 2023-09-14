package main

import (
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"time"
)

type Config interface {
	GetFilePath() string
}

var AutosaverNotRunningError = errors.New("autosaver is not running")

type AutosaverId uint32

type Autosaver struct {
	Config    Config
	Interval  time.Duration
	LastError error
	Running   bool

	id AutosaverId
}

var autosavers = make(map[AutosaverId]*bool)

func SaveConfig(config Config) error {
	cfgFile, err := os.OpenFile(config.GetFilePath(), os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if os.IsNotExist(err) {
		cfgFile, err = os.Create(config.GetFilePath())
	}

	if err != nil {
		return err
	}
	defer cfgFile.Close()

	content, err := json.Marshal(config)
	if err != nil {
		return err
	}
	_, err = cfgFile.Write(content)
	if err != nil {
		return err
	}

	return nil
}

func ReadConfig(config Config) error {
	cfgFile, err := os.OpenFile(config.GetFilePath(), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer cfgFile.Close()

	var fileSize int64
	fileStat, err := cfgFile.Stat()
	if err != nil {
		return err
	} else {
		fileSize = fileStat.Size()
	}

	fileContent := make([]byte, fileSize)
	_, err = cfgFile.Read(fileContent)

	err = json.Unmarshal(fileContent, config)
	if err != nil {
		return err
	}

	return nil
}

func NewAutosaver(config Config, interval time.Duration) *Autosaver {
	currentTimeUtc := time.Now().UTC()

	autosaver := new(Autosaver)
	autosaver.id = AutosaverId(rand.New(rand.NewSource(currentTimeUtc.UnixNano())).Uint32())
	autosaver.Interval = interval
	autosaver.Config = config

	return autosaver
}

func (a *Autosaver) Start() {
	stopAutosaver := false
	autosavers[a.id] = &stopAutosaver

	go func() {
		a.Running = true

		for {
			if stopAutosaver {
				a.Running = false
				return
			}

			err := SaveConfig(a.Config)
			if err != nil {
				a.LastError = err
				a.Running = false
				return
			}

			time.Sleep(a.Interval)
		}
	}()
}

func (a *Autosaver) Stop() error {
	if stopper, ok := autosavers[a.id]; ok {
		*stopper = true
	} else {
		return AutosaverNotRunningError
	}

	return nil
}
