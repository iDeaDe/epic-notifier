package epicgames

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func GetLink(slug string, categories []map[string]string) string {
	gameCategory := "p"

	for _, category := range categories {
		switch category["path"] {
		case "bundles":
			gameCategory = "bundles"
		}
	}

	slug = strings.ReplaceAll(slug, "/home", "")

	return fmt.Sprintf("%s%s/%s", GameLink, gameCategory, slug)
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

func GetGames() []RawGame {
	log.Println("Fetching new games from Epic Games API")
	resp, err := http.Get(EpicLink)
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
