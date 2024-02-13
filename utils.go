package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
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

func getWorkdir() string {
	resultWorkdir := os.Getenv("WORKDIR")

	if resultWorkdir == "" {
		var err error
		executable, err := os.Executable()
		if err != nil {
			resultWorkdir = "."
		} else {
			resultWorkdir = filepath.Dir(executable)
		}
	}

	return filepath.Clean(resultWorkdir)
}

func downloadFileByLink(link string) (io.ReadCloser, error) {
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func GetMonth(month time.Month) string {
	return Months[month-1]
}

func Add(number1 int, number2 int) int {
	return number1 + number2
}
