package epicgames

import (
	"github.com/ideade/epic-notifier/app/epicgames/graphql"
	"go.uber.org/zap"
	"net/http"
)

var logger *zap.Logger
var client *http.Client

func getLogger() *zap.Logger {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	return logger
}

func getClient() *http.Client {
	if client == nil {
		client = http.DefaultClient
	}

	return client
}

func SetLogger(newLogger *zap.Logger) {
	logger = newLogger
}

func SetClient(newClient *http.Client) {
	graphql.SetClient(newClient)
	client = newClient
}
