package main

import (
	"strconv"
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
