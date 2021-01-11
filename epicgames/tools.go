package epicgames

import (
	"encoding/json"
	"fmt"
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

var ProductType = []string{
	"product",
	"bundles",
}

func GetLink(slug string, categories []map[string]string) string {
	gameCategory := ProductType[0]

	// Такой вот костыль из-за того, что в url может быть как product, так и bundles
	for _, category := range categories {
		if category["path"] == ProductType[1] {
			gameCategory = ProductType[1]
		}
	}

	return fmt.Sprintf("%s%s/%s", GameLink, gameCategory, slug)
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
