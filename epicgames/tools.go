package epicgames

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"net/http"
)

var logger *zerolog.Logger
var client *http.Client

func getLogger() *zerolog.Logger {
	if logger == nil {
		newLogger := zerolog.New(zerolog.NewConsoleWriter())
		logger = &newLogger
	}

	return logger
}

func getClient() *http.Client {
	if client == nil {
		client = http.DefaultClient
	}

	return client
}

func SetLogger(newLogger *zerolog.Logger) {
	logger = newLogger
}

func SetClient(newClient *http.Client) {
	client = newClient
}

func getGameThumbnail(images []map[string]string) string {
	if len(images) == 0 {
		return ""
	}

	for _, image := range images {
		switch image["type"] {
		case
			"DieselStoreFrontTall",
			"Thumbnail",
			"VaultOpened",
			"DieselStoreFrontWide":
			return image["url"]
		}
	}

	return images[0]["url"]
}

func GetGames(link string) ([]RawGame, error) {
	getLogger().Debug().Msg("Fetching new games from Epic Games API")
	resp, err := getClient().Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	getLogger().Debug().Msg("Decoding JSON")
	responseData := new(Data)
	err = json.NewDecoder(resp.Body).Decode(responseData)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()

	rGames := responseData.Data.Catalog.SearchStore.Elements

	return rGames, nil
}
