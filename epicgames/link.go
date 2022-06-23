package epicgames

import (
	"fmt"
	"strings"
)

const fallbackLink = "https://t.me/epicgiveaways"
const baseLink = "https://store.epicgames.com/en-US/"

func GetLink(game *RawGame) string {
	slug := getSlug(game)

	gameCategory := "p"

	for _, category := range game.Categories {
		switch category["path"] {
		case "bundles":
			gameCategory = "bundles"
		}
	}

	slug = strings.ReplaceAll(slug, "/home", "")

	return fmt.Sprintf("%s%s/%s", baseLink, gameCategory, slug)
}

func getSlug(game *RawGame) string {
	for _, mapping := range game.OfferMappings {
		if mapping["pageType"] == "productHome" {
			return mapping["pageSlug"]
		}
	}

	for _, mapping := range game.CatalogNs.Mappings {
		if mapping["pageType"] == "productHome" {
			return mapping["pageSlug"]
		}
	}

	for _, item := range game.GameInfo {
		if item["key"] == "com.epicgames.app.productSlug" {
			return item["value"]
		}
	}

	if game.ProductSlug != "" && game.ProductSlug != "[]" {
		return game.ProductSlug
	}

	if game.UrlSlug != "" {
		return game.UrlSlug
	}

	return fallbackLink
}
