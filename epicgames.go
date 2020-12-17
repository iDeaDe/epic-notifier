package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"
)

const EpicLink = "https://store-site-backend-static.ak.epicgames.com/freeGamesPromotions?locale=ru-RU&country=RU&allowCountries=RU"
const GameLink = "https://www.epicgames.com/store/ru/"

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

type Giveaway struct {
	Games []Game
	Next  time.Time
}

type PromotionalOffer struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type RawGame struct {
	Title       string              `json:"title"`
	Image       []map[string]string `json:"keyImages"`
	GameInfo    []map[string]string `json:"customAttributes"`
	ProductSlug string              `json:"productSlug"`
	Categories  []map[string]string `json:"categories"`
	Price       struct {
		Total struct {
			Discount int `json:"discountPrice"`
			Original int `json:"originalPrice"`

			Currency struct {
				Decimals int `json:"decimals"`
			} `json:"currencyInfo"`
		} `json:"totalPrice"`
	} `json:"price"`
	Promotions struct {
		Current  []map[string][]PromotionalOffer `json:"promotionalOffers"`
		Upcoming []map[string][]PromotionalOffer `json:"upcomingPromotionalOffers"`
	} `json:"promotions"`
}

type Game struct {
	Title       string
	Publisher   string
	Developer   string
	IsAvailable bool
	Date        struct {
		Start time.Time
		End   time.Time
	}
	Image string
	Url   string
}

type Data struct {
	Data struct {
		Catalog struct {
			SearchStore struct {
				Elements []RawGame `json:"elements"`
			} `json:"searchStore"`
		} `json:"Catalog"`
	} `json:"data"`
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

func GetGiveaway() *Giveaway {
	var games []Game
	ga := new(Giveaway)

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

	// Выкладывать будем по московскому времени
	moscowLoc, _ := time.LoadLocation("Europe/Moscow")
	rGames := responseData.Data.Catalog.SearchStore.Elements

	// Собираем игры из ответа сервера
	log.Println("Selecting games we need to post")
	for _, rGame := range rGames {
		decimals := rGame.Price.Total.Currency.Decimals * 10
		discountPrice := rGame.Price.Total.Discount / decimals
		originalPrice := rGame.Price.Total.Original / decimals

		if !(discountPrice == 0 || originalPrice == 0) {
			continue
		}

		var localGameStruct = Game{}
		var dates PromotionalOffer

		localGameStruct.Title = rGame.Title // Название игры

		// Находим даты начала и окончания раздачи
		if len(rGame.Promotions.Current) > 0 {
			/**
			Эпик продолжает вставлять палки в колёса
			Пришлось добавить ещё и такую проверку
			Что же будет дальше?
			*/
			if rGame.Promotions.Current == nil {
				continue
			}

			localGameStruct.IsAvailable = true
			dates = rGame.Promotions.Current[0]["promotionalOffers"][0]
		} else {
			if rGame.Promotions.Upcoming == nil {
				continue
			}

			localGameStruct.IsAvailable = false
			dates = rGame.Promotions.Upcoming[0]["promotionalOffers"][0]
		}

		// Парсим даты по московскому времени
		localGameStruct.Date.Start, _ = time.ParseInLocation(
			time.RFC3339,
			dates.StartDate,
			moscowLoc)
		localGameStruct.Date.End, _ = time.ParseInLocation(
			time.RFC3339,
			dates.EndDate,
			moscowLoc)

		if !localGameStruct.IsAvailable {
			// Устанавливаем время до следующей раздачи, а если находим раньше текущего - перезаписываем
			if ga.Next.IsZero() || localGameStruct.Date.Start.Before(ga.Next) {
				ga.Next = localGameStruct.Date.Start
			}

			continue
		}

		/**
		Этот кусочек добавлен после того, как эпики в ответе стали выдавать игру 2018 года.
		Посмотрел ответ сервера, там её больше нет, но осадочек остался
		*/
		if localGameStruct.Date.Start.Before(time.Now().AddDate(0, -2, 0)) {
			continue
		}

		// Ищем обложку игры
		sort.Slice(rGame.Image, func(i, _ int) bool {
			return rGame.Image[i]["type"] == "Thumbnail"
		})

		localGameStruct.Image = rGame.Image[0]["url"]                      // Обложка
		localGameStruct.Url = GetLink(rGame.ProductSlug, rGame.Categories) // Ссылка на страницу игры

		// Данный массив может меняться, поэтому ищем нужную информацию таким способом
		for _, gameInfo := range rGame.GameInfo {
			fieldVal := gameInfo["value"]

			// Тут любое из полей может быть пустым
			switch gameInfo["key"] {
			// Разработчик
			case "developerName":
				localGameStruct.Developer = fieldVal
				break
			// Издатель
			case "publisherName":
				localGameStruct.Publisher = fieldVal
				break
			}
		}

		games = append(games, localGameStruct)
	}

	ga.Games = games

	return ga
}
