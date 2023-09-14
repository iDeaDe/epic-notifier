package epicgames

import (
	"encoding/json"
	"log"
	"net/http"
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

var logger *log.Logger

func getLogger() *log.Logger {
	if logger == nil {
		logger = log.Default()
	}

	return logger
}

func SetLogger(newLogger *log.Logger) {
	logger = newLogger
}

func GetGameThumbnail(images []map[string]string) string {
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

func GetMonth(month time.Month) string {
	return Months[month-1]
}

func GetGames(link string) ([]RawGame, error) {
	getLogger().Println("Fetching new games from Epic Games API")
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	getLogger().Println("Decoding JSON")
	responseData := new(Data)
	err = json.NewDecoder(resp.Body).Decode(responseData)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()

	rGames := responseData.Data.Catalog.SearchStore.Elements

	return rGames, nil
}
