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

func GetGameThumbnail(images []map[string]string) string {
	if len(images) == 0 {
		return ""
	}

	for _, image := range images {
		switch image["type"] {
		case
			"DieselStoreFrontTall",
			"Thumbnail",
			"VaultOpened":
			return image["url"]
		}
	}

	return images[0]["url"]
}

func GetMonth(month time.Month) string {
	return Months[month-1]
}

func GetGames(link string) []RawGame {
	log.Println("Fetching new games from Epic Games API")
	resp, err := http.Get(link)
	if err != nil {
		log.Panicln(err)
	}
	defer resp.Body.Close()

	log.Println("Decoding JSON")
	responseData := new(Data)
	err = json.NewDecoder(resp.Body).Decode(responseData)
	if err != nil {
		log.Panicln(err)
	}
	_ = resp.Body.Close()

	rGames := responseData.Data.Catalog.SearchStore.Elements

	return rGames
}
