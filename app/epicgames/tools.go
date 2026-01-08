package epicgames

import (
	"encoding/json"
	"errors"
	"github.com/ideade/epic-notifier/app/epicgames/graphql"
	"math"
)

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
			"DieselStoreFrontWide",
			"GalleryImage":
			return image["url"]
		}
	}

	return images[0]["url"]
}

func getGames(link string) ([]rawGame, error) {
	getLogger().Debug("Fetching new games from Epic Games API")
	resp, err := getClient().Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	getLogger().Debug("Decoding JSON")
	responseData := new(data)
	err = json.NewDecoder(resp.Body).Decode(responseData)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()

	rGames := responseData.Data.Catalog.SearchStore.Elements

	return rGames, nil
}

func convertPrice(offerPrice price) gamePrice {
	decimals := math.Pow(10, offerPrice.Total.Currency.Decimals)
	originalPrice := 0.0
	if decimals > 0 {
		originalPrice = offerPrice.Total.Original / decimals
	}

	return gamePrice{
		Original: originalPrice,
		Format:   offerPrice.Total.FormatPrice.OriginalPrice,
		Currency: offerPrice.Total.CurrencyCode,
	}
}

func fillGameDetails(locale, country string, game *Game) error {
	mapping, err := graphql.GetMappingByPageSlug(locale, game.Slug)
	if err != nil {
		return err
	}

	offerId := mapping.Mappings.OfferId
	if offerId == "" {
		offerId = game.Id
	}

	catalogOffer, err := graphql.GetCatalogOffer(locale, country, offerId, mapping.SandboxId)
	if err != nil {
		return err
	}

	if catalogOffer == nil {
		var productInfo *StoreProductInfo
		productInfo, err = FetchStoreProductInfo(game.Slug)
		if err != nil {
			return err
		}

		for _, page := range productInfo.Pages {
			if page.Offer != nil && page.Offer.HasOffer && page.Offer.Id != "" {
				offerId = page.Offer.Id
			}
		}

		catalogOffer, err = graphql.GetCatalogOffer(locale, country, offerId, mapping.SandboxId)
		if err != nil {
			return err
		}
	}

	if catalogOffer == nil {
		return errors.New("failed to fetch detailed offer info")
	}

	var platforms []string
	var genres []string

	for _, tag := range catalogOffer.Tags {
		switch tag.GroupName {
		case "platform":
			platforms = append(platforms, tag.Name)
		case "genre":
			genres = append(genres, tag.Name)
		}
	}

	if len(catalogOffer.Description) > 20 && catalogOffer.Title != catalogOffer.Description {
		game.Description = catalogOffer.Description
	}

	if game.Developer == "" {
		game.Developer = catalogOffer.DeveloperDisplayName
	}

	if game.Publisher == "" {
		game.Publisher = catalogOffer.PublisherDisplayName
	}

	game.Price = convertPrice(catalogOffer.Price)
	game.CountriesBlacklist = catalogOffer.CountriesBlacklist
	game.Platforms = platforms
	game.Genres = genres

	return nil
}
