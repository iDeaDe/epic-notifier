package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ideade/epic-notifier/app/currency"
	"github.com/ideade/epic-notifier/app/epicgames"
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

var currencyToSymbol = map[string]string{
	"RUB": "₽",
	"KZT": "₸",
	"USD": "$",
}

var currencyUpdater currency.Updater

func GetMonth(month time.Month) string {
	return Months[month-1]
}

func Add(number1 int, number2 int) int {
	return number1 + number2
}

func FormatMoney(game epicgames.Game, target string) string {
	convertedMoney := strconv.FormatFloat(
		currencyUpdater.Convert(game.Price.Original, *currency.NewPair(game.Price.Currency, target)),
		'f',
		2,
		64,
	)

	if _, ok := currencyToSymbol[target]; !ok {
		return fmt.Sprintf("%s %s", convertedMoney, target)
	}

	if target == "USD" {
		return fmt.Sprintf("$%s", convertedMoney)
	} else {
		return fmt.Sprintf("%s%s", convertedMoney, currencyToSymbol[target])
	}
}

func Platforms(game epicgames.Game) string {
	platforms := game.Platforms
	if len(platforms) == 0 {
		return "Неизвестно"
	}

	return strings.Join(platforms, ", ")
}

func Join(items []string) string {
	return strings.Join(items, ", ")
}

func BlacklistedCountries(items []string) string {
	var countryNames []string
	var countriesToPrint []string
	withMoreItemsText := false

	if len(items) > 4 {
		countryNames = make([]string, 4)
		countriesToPrint = items[:3]
		withMoreItemsText = true
	} else {
		countryNames = make([]string, len(items))
		countriesToPrint = items
	}

	for index, countryCode := range countriesToPrint {
		countryNames[index] = countriesMap[countryCode]
	}

	result := strings.Join(countryNames, ", ")

	if withMoreItemsText {
		result = result + " и ещё " + strconv.FormatInt(int64(len(items)-4), 10)
	}

	return result
}
