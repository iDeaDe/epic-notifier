package main

import (
	"bytes"
	"github.com/ideade/epic-notifier/epicgames"
	"github.com/ideade/epic-notifier/telegram"
	"github.com/rs/zerolog"
	"html/template"
	"path/filepath"
	"strconv"
)

var logger *zerolog.Logger

func Logger() *zerolog.Logger {
	if logger == nil {
		instance := zerolog.New(zerolog.NewConsoleWriter())
		logger = &instance
	}

	return logger
}

func PostGame(client *telegram.Client, chatId string, game *epicgames.Game) (string, error) {
	tpl, err := template.New("game.gohtml").
		Funcs(
			template.FuncMap{
				"month": epicgames.GetMonth,
			},
		).
		ParseFiles(filepath.Join("template", "game.gohtml"))

	if err != nil {
		return "", err
	}

	buffer := bytes.Buffer{}
	if err = tpl.Execute(&buffer, game); err != nil {
		return "", err
	}

	photo, err := downloadFileByLink(game.Image)
	if err != nil {
		return "", err
	}

	request := telegram.SendPhotoRequest{
		Photo:               photo,
		ChatId:              chatId,
		Caption:             buffer.String(),
		ParseMode:           "HTML",
		HasSpoiler:          false,
		DisableNotification: false,
		ProtectContent:      false,
	}

	message, err := client.SendPhoto(&request)
	if err != nil {
		return "", err
	}

	return strconv.FormatUint(message.MessageId, 10), nil
}
