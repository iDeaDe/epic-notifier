package logging

import (
	"fmt"
	"github.com/ideade/epic-notifier/app/telegram"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func AddNotificationHook(logger *zap.Logger, telegramClient *telegram.Client, chatId string) *zap.Logger {
	return logger.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
		var err error = nil

		if entry.Level > zapcore.WarnLevel {
			_, err = telegramClient.SendMessage(&telegram.SendMessageRequest{
				ChatId: chatId,
				Text: fmt.Sprintf(
					"*%s:* %s\n\n*Caller:* %s\n\n*Stacktrace:*\n```\n%s\n```",
					telegram.EscapeString(entry.Level.CapitalString(), telegram.ParseModeMarkdown),
					telegram.EscapeString(entry.Message, telegram.ParseModeMarkdown),
					telegram.EscapeString(entry.Caller.String(), telegram.ParseModeMarkdown),
					telegram.EscapeString(entry.Stack, telegram.ParseModeMarkdown)),
				ParseMode: telegram.ParseModeMarkdown,
			})
		}

		return err
	}))
}
