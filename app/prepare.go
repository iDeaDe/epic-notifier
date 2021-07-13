package app

import (
	"github.com/ideade/epic-notifier/telegram"
	"os"
	"path"
	"strconv"
)

func Prepare(params Parameters) *App {
	instance := new(App)
	instance.params = params
	config := &instance.config

	/*
		Смена директории для того, чтобы конфиг создавался рядом с исполняемым файлом
	*/
	executable, err := os.Executable()
	if err != nil {
		logger.Panicln(err)
	}
	execPath := path.Dir(executable)
	err = os.Chdir(execPath)
	if err != nil {
		logger.Panicln(err)
	}

	config.Name = params.ConfigFile
	config.GetConfig()

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		logger.Panicln("TELEGRAM_TOKEN not found")
	}

	instance.tg = new(telegram.TelegramSettings)
	instance.tg.Token = telegramToken
	instance.tg.ChannelName = config.Content.Channel
	if params.TestChannel != "" {
		instance.tg.ChannelName = params.TestChannel
	}

	/*
		Пересоздание поста с напоминанием о последнем дне раздаче.
	*/
	if params.Remind {
		if config.Content.RemindPostId == 0 {
			logger.Println("Remind post does not exist")
		} else {
			err = instance.tg.RemoveRemind(strconv.FormatUint(config.Content.RemindPostId, 10))
			if err != nil {
				logger.Println(err)
			}

			config.Content.RemindPostId = 0
			_ = config.SaveConfig()
		}
	}

	instance.remindSent = false
	if config.Content.RemindPostId != 0 {
		instance.remindSent = true
	}

	/*
		Пересоздание поста с анонсом следующей раздачи.
		Это не работает. В Телеграме нельзя ботом удалять посты старше 48 часов.
	*/
	if params.Next {
		if config.Content.NextPostId == 0 {
			logger.Println("Next giveaway post does not exist")
		} else {
			err = instance.tg.RemoveNextPost(strconv.FormatUint(config.Content.NextPostId, 10))
			if err != nil {
				logger.Println(err)
			}

			config.Content.NextPostId = 0
			_ = config.SaveConfig()
		}
	}

	return instance
}
