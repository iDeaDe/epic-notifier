package epicgames

import (
	"log"
	"time"
)

const EpicLink = "https://store-site-backend-static.ak.epicgames.com/freeGamesPromotions?locale=ru-RU&country=RU&allowCountries=RU"
const GameLink = "https://www.epicgames.com/store/ru/"

type Giveaway struct {
	CurrentGames []Game
	NextGames    []Game
	Next         time.Time
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
	Price       struct {
		Discount int
		Original int
	}
	Date struct {
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

func GetGiveaway() *Giveaway {
	var currentGames []Game
	var nextGames []Game
	ga := new(Giveaway)

	// Выкладывать будем по московскому времени
	moscowLoc, _ := time.LoadLocation("Europe/Moscow")
	rGames := GetGames()

	// Собираем игры из ответа сервера
	log.Println("Converting raw information to structures")
	for _, rGame := range rGames {
		decimals := rGame.Price.Total.Currency.Decimals * 10
		discountPrice := rGame.Price.Total.Discount / decimals
		originalPrice := rGame.Price.Total.Original / decimals

		var localGameStruct = Game{}
		var dates PromotionalOffer

		localGameStruct.Title = rGame.Title // Название игры

		if len(rGame.Promotions.Current) == 0 && len(rGame.Promotions.Upcoming) == 0 {
			continue
		}

		// Находим даты начала и окончания раздачи
		if len(rGame.Promotions.Current) > 0 {
			/*
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

		if !(discountPrice == 0 || originalPrice == 0) && localGameStruct.IsAvailable {
			continue
		}

		if !localGameStruct.IsAvailable {
			// Устанавливаем время до следующей раздачи, а если находим раньше текущего - перезаписываем
			if ga.Next.IsZero() || localGameStruct.Date.Start.Before(ga.Next) {
				ga.Next = localGameStruct.Date.Start
			}
		}

		/**
		Этот кусочек добавлен после того, как эпики в ответе стали выдавать игру 2018 года.
		Посмотрел ответ сервера, там её больше нет, но осадочек остался
		*/
		if localGameStruct.Date.Start.Before(time.Now().AddDate(0, -2, 0)) {
			continue
		}

		// Ищем обложку игры
		for _, image := range rGame.Image {
			// Для тайных игр
			if !localGameStruct.IsAvailable &&
				len(rGame.Image) == 1 &&
				image["type"] == "VaultClosed" {
				localGameStruct.Image = image["url"]
				break
			}

			if image["type"] == "DieselStoreFrontTall" {
				localGameStruct.Image = image["url"]
				break
			}
		}

		if localGameStruct.Image == "" && len(rGame.Image) == 1 {
			localGameStruct.Image = rGame.Image[0]["url"]
		}

		if rGame.ProductSlug == "[]" {
			localGameStruct.Url = "https://t.me/epicgiveaways"
		} else {
			localGameStruct.Url = GetLink(rGame.ProductSlug, rGame.Categories) // Ссылка на страницу игры
		}

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

		if localGameStruct.IsAvailable {
			currentGames = append(currentGames, localGameStruct)
		} else {
			nextGames = append(nextGames, localGameStruct)
		}
	}

	ga.CurrentGames = currentGames
	ga.NextGames = nextGames

	return ga
}
