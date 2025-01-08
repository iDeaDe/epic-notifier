package main

import (
	"github.com/ideade/epic-notifier/app/currency"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
	"time"
)

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
	resp, err := GlobalHttpClient().Get(link)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func sleep(logger *zap.Logger, duration time.Duration) {
	logger.Info("going to sleep till " + time.Now().Add(duration).Format(time.RFC1123))
	time.Sleep(duration)
}

type NopCurrencyUpdater struct{}

func (ncu *NopCurrencyUpdater) SetErrorHandler(func(error)) {}
func (ncu *NopCurrencyUpdater) Convert(sum float64, _ currency.Pair) float64 {
	return sum
}
func (ncu *NopCurrencyUpdater) AddPair(currency.Pair) {}
func (ncu *NopCurrencyUpdater) Update()               {}
