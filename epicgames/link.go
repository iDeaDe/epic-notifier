package epicgames

import (
	"fmt"
	"strings"
)

const fallbackLink = "https://t.me/epicgiveaways"
const baseLink = "https://store.epicgames.com/en-US/"

var slugPageTypes = []string{
	"productHome",
	"addon--cms-hybrid",
}

func getLink(game *RawGame) string {
	slug := getSlug(game)

	if slug == "" {
		return fallbackLink
	}

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
	/*
		Есть ощущение, что offerMappings и catalogNs.mappings - одно и то же, но лучше чекать оба
	*/
	for _, mapping := range game.OfferMappings {
		if inArray(mapping["pageType"], slugPageTypes) && isNotEmpty(mapping["pageSlug"]) {
			return mapping["pageSlug"]
		}
	}

	for _, mapping := range game.CatalogNs.Mappings {
		if mapping["pageType"] == "productHome" && isNotEmpty(mapping["pageSlug"]) {
			return mapping["pageSlug"]
		}
	}

	for _, item := range game.GameInfo {
		if item["key"] == "com.epicgames.app.productSlug" && isNotEmpty(item["value"]) {
			return item["value"]
		}
	}

	if isNotEmpty(game.ProductSlug) {
		return game.ProductSlug
	}

	if isNotEmpty(game.UrlSlug) {
		return game.UrlSlug
	}

	return ""
}

func isNotEmpty(slug string) bool {
	return slug != "" && slug != "[]"
}

func inArray[E comparable](value E, slice []E) bool {
	for _, sliceItem := range slice {
		if value == sliceItem {
			return true
		}
	}

	return false
}
