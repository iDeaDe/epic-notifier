package main

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ideade/epic-notifier/epicgames"
	"github.com/ideade/epic-notifier/telegram"
	"github.com/rs/zerolog"
)

var logger *zerolog.Logger

func Logger() *zerolog.Logger {
	if logger == nil {
		instance := zerolog.New(zerolog.NewConsoleWriter())
		logger = &instance
	}

	return logger
}

func AddLoggerHook(hook zerolog.Hook) {
	Logger()

	newLogger := logger.Hook(hook)
	logger = &newLogger
}

type NotifyHook struct {
	telegramClient *telegram.Client
	chatId         string
}

func NewNotifyHook(telegramClient *telegram.Client, chatId string) *NotifyHook {
	return &NotifyHook{telegramClient, chatId}
}

func (notifyHook *NotifyHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level > zerolog.WarnLevel {
		message := msg
		if message == "" {
			message = "error in app! check logs"
		}

		_, _ = notifyHook.telegramClient.SendMessage(&telegram.SendMessageRequest{
			ChatId: notifyHook.chatId,
			Text:   fmt.Sprintf("[%s] %s", level.String(), message),
		})
	}
}

type Poster struct {
	telegramClient *telegram.Client
	Timezone       *time.Location
	ChatId         string
	templateDir    string
	silentMode     bool
}

func NewPoster(client *telegram.Client, chatId string) *Poster {
	return &Poster{
		telegramClient: client,
		Timezone:       time.UTC,
		ChatId:         chatId,
		templateDir:    "template",
		silentMode:     false,
	}
}

func (poster *Poster) SetTimezone(tz string) error {
	var err error
	poster.Timezone, err = time.LoadLocation(tz)
	return err
}

func (poster *Poster) SetTemplateDir(dir string) {
	poster.templateDir = dir
}

func (poster *Poster) SetSilentMode(silent bool) {
	poster.silentMode = silent
}

type PostGameTemplateData struct {
	Game             *epicgames.Game
	AvailabilityTime *time.Time
}

func (poster *Poster) PostCurrentGames(games []epicgames.Game) ([]string, error) {
	tpl, err := template.New("game.gohtml").
		Funcs(
			template.FuncMap{
				"month": GetMonth,
			},
		).
		ParseFiles(filepath.Join(poster.templateDir, "game.gohtml"))

	if err != nil {
		return nil, err
	}

	postIds := make([]string, len(games))

	for _, game := range games {
		endDate := game.Date.End.In(poster.Timezone)

		data := &PostGameTemplateData{
			Game:             &game,
			AvailabilityTime: &endDate,
		}

		buffer := bytes.Buffer{}
		if err = tpl.Execute(&buffer, data); err != nil {
			return postIds, err
		}

		photo, err := downloadFileByLink(game.Image)
		if err != nil {
			return postIds, err
		}

		request := telegram.SendPhotoRequest{
			Photo:               photo,
			ChatId:              poster.ChatId,
			Caption:             buffer.String(),
			ParseMode:           telegram.ParseModeHtml,
			HasSpoiler:          false,
			DisableNotification: poster.silentMode,
			ProtectContent:      false,
		}

		message, err := poster.telegramClient.SendPhoto(&request)
		if err != nil {
			return postIds, err
		}

		postIds = append(postIds, strconv.FormatUint(message.MessageId, 10))
	}

	return postIds, nil
}

type PostAnnounceTemplateData struct {
	Giveaway  *epicgames.Giveaway
	StartDate *time.Time
}

func (poster *Poster) PostAnnounce(giveaway *epicgames.Giveaway) ([]string, error) {
	tpl, err := template.New("announce.gohtml").
		Funcs(
			template.FuncMap{
				"month": GetMonth,
				"add":   Add,
			},
		).
		ParseFiles(filepath.Join(poster.templateDir, "announce.gohtml"))

	if err != nil {
		return nil, err
	}

	startDate := giveaway.Next.In(poster.Timezone)

	data := &PostAnnounceTemplateData{
		Giveaway:  giveaway,
		StartDate: &startDate,
	}

	buffer := bytes.Buffer{}
	if err = tpl.Execute(&buffer, data); err != nil {
		return nil, err
	}

	var media []telegram.InputMediaPhoto

	for index, game := range giveaway.NextGames {
		photo, err := downloadFileByLink(game.Image)
		if err != nil {
			return nil, err
		}

		if index == 0 {
			media = append(media, telegram.InputMediaPhoto{
				Media:     photo,
				Caption:   buffer.String(),
				ParseMode: telegram.ParseModeHtml,
			})
		} else {
			media = append(media, telegram.InputMediaPhoto{Media: photo})
		}
	}

	request := telegram.SendMediaGroupRequest{
		Media:               media,
		ChatId:              poster.ChatId,
		DisableNotification: poster.silentMode,
		ProtectContent:      false,
	}

	messages, err := poster.telegramClient.SendMediaGroup(&request)
	if err != nil {
		return nil, err
	}

	postIds := make([]string, len(messages))

	for _, message := range messages {
		postIds = append(postIds, strconv.FormatUint(message.MessageId, 10))
	}

	postIds = append(postIds)

	return postIds, nil
}

func (poster *Poster) PostRemind() (string, error) {
	tpl, err := template.New("remind.gohtml").
		Funcs(
			template.FuncMap{
				"month": GetMonth,
			},
		).
		ParseFiles(filepath.Join(poster.templateDir, "remind.gohtml"))

	if err != nil {
		return "", err
	}

	buffer := bytes.Buffer{}
	if err = tpl.Execute(&buffer, nil); err != nil {
		return "", err
	}

	message, err := poster.telegramClient.SendMessage(&telegram.SendMessageRequest{
		ChatId:              poster.ChatId,
		Text:                buffer.String(),
		ParseMode:           telegram.ParseModeHtml,
		DisableNotification: poster.silentMode,
		ProtectContent:      false,
	})

	if err != nil {
		return "", err
	}

	return strconv.FormatUint(message.MessageId, 10), nil
}
