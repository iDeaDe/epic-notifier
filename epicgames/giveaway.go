package epicgames

import (
	"math"
	"time"
)

const epicLink = "https://store-site-backend-static-ipv4.ak.epicgames.com/freeGamesPromotions?locale=ru-RU&country=KZ&allowCountries=KZ"
const epicRuLink = "https://store-site-backend-static-ipv4.ak.epicgames.com/freeGamesPromotions?locale=ru-RU&country=RU&allowCountries=RU"

type Giveaway struct {
	CurrentGames []Game
	NextGames    []Game
	Next         time.Time
}

type PromotionalOffer struct {
	StartDate       string `json:"startDate"`
	EndDate         string `json:"endDate"`
	DiscountSetting struct {
		DiscountType       string `json:"discountType"`
		DiscountPercentage int    `json:"discountPercentage"`
	} `json:"discountSetting"`
}

type Promotion struct {
	PromotionalOffers []PromotionalOffer `json:"promotionalOffers"`
}
type Promotions []Promotion

type RawGame struct {
	Id          string              `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Images      []map[string]string `json:"keyImages"`
	GameInfo    []map[string]string `json:"customAttributes"`
	UrlSlug     string              `json:"urlSlug"`
	ProductSlug string              `json:"productSlug"`
	Categories  []map[string]string `json:"categories"`
	CatalogNs   struct {
		Mappings []map[string]string `json:"mappings"`
	} `json:"catalogNs"`
	OfferMappings []map[string]string `json:"offerMappings"`
	Price         struct {
		Total struct {
			Discount float64 `json:"discountPrice"`
			Original float64 `json:"originalPrice"`

			CurrencyCode string `json:"currencyCode"`
			Currency     struct {
				Decimals float64 `json:"decimals"`
			} `json:"currencyInfo"`

			FormatPrice struct {
				OriginalPrice string `json:"originalPrice"`
			} `json:"fmtPrice"`
		} `json:"totalPrice"`
	} `json:"price"`
	Promotions struct {
		Current  Promotions `json:"promotionalOffers"`
		Upcoming Promotions `json:"upcomingPromotionalOffers"`
	} `json:"promotions"`
}

type Game struct {
	Id          string
	Title       string
	Description string
	Publisher   string
	Developer   string
	Type        GameType
	Price       struct {
		Original float64
		Format   string
		Currency string
	}
	Date struct {
		Start time.Time
		End   time.Time
	}
	Image       string
	Url         string
	AvailableRu bool
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

func getGiveaway(url string) (*Giveaway, error) {
	ga := new(Giveaway)

	rGames, err := GetGames(url)
	if err != nil {
		return nil, err
	}

	ga.CurrentGames = []Game{}
	ga.NextGames = []Game{}

	// Собираем игры из ответа сервера
	getLogger().Debug().Msg("Converting raw information to structures")
	for _, rGame := range rGames {
		decimals := math.Pow(10, rGame.Price.Total.Currency.Decimals)
		discountPrice := 0.0
		originalPrice := 0.0
		if decimals > 0 {
			discountPrice = rGame.Price.Total.Discount / decimals
			originalPrice = rGame.Price.Total.Original / decimals
		}

		var localGameStruct = Game{}

		localGameStruct.Id = rGame.Id
		localGameStruct.Title = rGame.Title

		if len(rGame.Description) > 20 && rGame.Title != rGame.Description {
			localGameStruct.Description = rGame.Description
		}

		if rGame.Promotions.Current == nil {
			rGame.Promotions.Current = Promotions{}
		}

		if rGame.Promotions.Upcoming == nil {
			rGame.Promotions.Upcoming = Promotions{}
		}

		var dates *PromotionalOffer
		localGameStruct.Type, dates = GetType(&rGame)

		if localGameStruct.Type == UnknownType {
			continue
		}

		// Парсим даты по московскому времени
		localGameStruct.Date.Start, _ = time.Parse(DateTimeFormat, dates.StartDate)
		localGameStruct.Date.End, _ = time.Parse(DateTimeFormat, dates.EndDate)

		if !(discountPrice == 0 || originalPrice == 0) && localGameStruct.Type == Current {
			continue
		}

		if localGameStruct.Type == Upcoming {
			// Устанавливаем время до следующей раздачи, а если находим раньше текущего - перезаписываем
			if ga.Next.IsZero() || localGameStruct.Date.Start.Before(ga.Next) {
				ga.Next = localGameStruct.Date.Start
			}
		}

		localGameStruct.Image = getGameThumbnail(rGame.Images)
		localGameStruct.Price.Original = originalPrice
		localGameStruct.Price.Format = rGame.Price.Total.FormatPrice.OriginalPrice
		localGameStruct.Price.Currency = rGame.Price.Total.CurrencyCode

		localGameStruct.Url = getLink(&rGame)

		// Данный массив может меняться, поэтому ищем нужную информацию таким способом
		for _, gameInfo := range rGame.GameInfo {
			fieldVal := gameInfo["value"]

			// Тут любое из полей может быть пустым
			switch gameInfo["key"] {
			// Разработчик
			case "developerName":
				localGameStruct.Developer = fieldVal
			// Издатель
			case "publisherName":
				localGameStruct.Publisher = fieldVal
			}
		}

		switch localGameStruct.Type {
		case Current:
			ga.CurrentGames = append(ga.CurrentGames, localGameStruct)
		case Upcoming:
			ga.NextGames = append(ga.NextGames, localGameStruct)
		}
	}

	filterNextGames(ga)

	return ga, nil
}

func GetGiveaway() (*Giveaway, error) {
	globalGiveaway, err := getGiveaway(epicLink)
	if err != nil {
		return nil, err
	}

	ruGiveaway, err := getGiveaway(epicRuLink)
	if err != nil {
		return globalGiveaway, err
	}

	nameToGameMap := map[string]*Game{}

	for _, game := range ruGiveaway.CurrentGames {
		nameToGameMap[game.Id] = &game
	}

	for index, game := range globalGiveaway.CurrentGames {
		ruGame, ok := nameToGameMap[game.Id]
		globalGiveaway.CurrentGames[index].AvailableRu = ok

		if ok {
			globalGiveaway.CurrentGames[index].Price = ruGame.Price
		}
	}

	nameToGameMap = map[string]*Game{}

	for _, game := range ruGiveaway.NextGames {
		nameToGameMap[game.Id] = &game
	}

	for index, game := range globalGiveaway.NextGames {
		_, ok := nameToGameMap[game.Id]
		globalGiveaway.NextGames[index].AvailableRu = ok
	}

	return globalGiveaway, nil
}
