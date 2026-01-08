package epicgames

import (
	"errors"
	"math"
	"net/url"
	"strings"
	"time"
)

const epicLink = "https://store-site-backend-static-ipv4.ak.epicgames.com/freeGamesPromotions"

type Giveaway struct {
	CurrentGames []Game
	NextGames    []Game
	Next         time.Time
}

type Game struct {
	Id          string
	Namespace   string
	Title       string
	Description string
	Publisher   string
	Developer   string
	Price       gamePrice
	Date        struct {
		Start time.Time
		End   time.Time
	}
	Image              string
	Url                string
	Slug               string
	Platforms          []string
	CountriesBlacklist []string
	Genres             []string

	gameType gameType
}

type gamePrice struct {
	Original float64
	Format   string
	Currency string
}

type promotionalOffer struct {
	StartDate       string `json:"startDate"`
	EndDate         string `json:"endDate"`
	DiscountSetting struct {
		DiscountType       string `json:"discountType"`
		DiscountPercentage int    `json:"discountPercentage"`
	} `json:"discountSetting"`
}

type promotion struct {
	PromotionalOffers []promotionalOffer `json:"promotionalOffers"`
}
type promotions []promotion

type price struct {
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
}

type rawGame struct {
	Id          string              `json:"id"`
	Namespace   string              `json:"namespace"`
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
	Price         price               `json:"price"`
	Promotions    struct {
		Current  promotions `json:"promotionalOffers"`
		Upcoming promotions `json:"upcomingPromotionalOffers"`
	} `json:"promotions"`
}

type data struct {
	Data struct {
		Catalog struct {
			SearchStore struct {
				Elements []rawGame `json:"elements"`
			} `json:"searchStore"`
		} `json:"Catalog"`
	} `json:"data"`
}

func GetGiveaway(locale, country string) (*Giveaway, error) {
	ga := new(Giveaway)

	freeGamesUrl, _ := url.Parse(epicLink)
	queryParams := freeGamesUrl.Query()
	queryParams.Set("locale", locale)
	queryParams.Set("country", country)
	queryParams.Set("allowCountries", country)
	freeGamesUrl.RawQuery = queryParams.Encode()

	rGames, err := getGames(freeGamesUrl.String())
	if err != nil {
		return nil, err
	}

	ga.CurrentGames = []Game{}
	ga.NextGames = []Game{}

	// Собираем игры из ответа сервера
	getLogger().Debug("Converting raw information to structures")
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
		localGameStruct.Namespace = rGame.Namespace
		localGameStruct.Title = rGame.Title
		localGameStruct.Slug = getSlug(&rGame)

		if len(rGame.Description) > 20 && rGame.Title != rGame.Description {
			localGameStruct.Description = rGame.Description
		}

		if rGame.Promotions.Current == nil {
			rGame.Promotions.Current = promotions{}
		}

		if rGame.Promotions.Upcoming == nil {
			rGame.Promotions.Upcoming = promotions{}
		}

		var dates *promotionalOffer
		localGameStruct.gameType, dates = getType(&rGame)

		if localGameStruct.gameType == gameTypeUnknown {
			continue
		}

		// Парсим даты по московскому времени
		localGameStruct.Date.Start, _ = time.Parse(dateTimeFormat, dates.StartDate)
		localGameStruct.Date.End, _ = time.Parse(dateTimeFormat, dates.EndDate)

		if !(discountPrice == 0 || originalPrice == 0) && localGameStruct.gameType == gameTypeCurrent {
			continue
		}

		if localGameStruct.gameType == gameTypeUpcoming {
			// Устанавливаем время до следующей раздачи, а если находим раньше текущего - перезаписываем
			if ga.Next.IsZero() || localGameStruct.Date.Start.Before(ga.Next) {
				ga.Next = localGameStruct.Date.Start
			}
		}

		if localGameStruct.gameType == gameTypeCurrent && ga.Next.IsZero() {
			ga.Next = localGameStruct.Date.End
		}

		localGameStruct.Image = getGameThumbnail(rGame.Images)
		localGameStruct.Price = convertPrice(rGame.Price)
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

		if localGameStruct.gameType == gameTypeCurrent && !strings.Contains(localGameStruct.Url, "/bundles/") {
			if err = fillGameDetails(locale, country, &localGameStruct); err != nil {
				getLogger().Error(err.Error())
			}
		}

		switch localGameStruct.gameType {
		case gameTypeCurrent:
			ga.CurrentGames = append(ga.CurrentGames, localGameStruct)
		case gameTypeUpcoming:
			ga.NextGames = append(ga.NextGames, localGameStruct)
		}
	}

	filterNextGames(ga)

	if len(ga.CurrentGames) == 0 {
		return nil, errors.New("incorrect response from Epic Games")
	}

	return ga, nil
}
