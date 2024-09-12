package main

import (
	"github.com/ideade/epic-notifier/epicgames"
	"strconv"
	"strings"
	"time"
)

var Months = []string{
	"января",
	"февраля",
	"марта",
	"апреля",
	"мая",
	"июня",
	"июля",
	"августа",
	"сентября",
	"октября",
	"ноября",
	"декабря",
}

func GetMonth(month time.Month) string {
	return Months[month-1]
}

func Add(number1 int, number2 int) int {
	return number1 + number2
}

func FormatMoney(sum float64) string {
	return strconv.FormatFloat(sum, 'f', 2, 64)
}

func Platforms(game epicgames.Game) string {
	offer, err := epicgames.GetCatalogOffer(game.Namespace, game.Id, "ru-RU")
	if err != nil {
		Logger().Error().Err(err).Send()

		return "Неизвестно"
	}

	platforms := offer.GetTagsByGroupName("platform")
	if len(platforms) == 0 {
		return "Неизвестно"
	}

	builder := strings.Builder{}
	builder.WriteString(platforms[0].Name)

	for _, platform := range platforms[1:] {
		builder.WriteString(", ")
		builder.WriteString(platform.Name)
	}

	return builder.String()
}
