package epicgames

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

const QUERY = `
query getCatalogOffer($sandboxId: String!, $offerId: String!, $locale: String) {
  Catalog {
    catalogOffer(namespace: $sandboxId, id: $offerId, locale: $locale) {
      title
      id
      namespace
      countriesBlacklist
      countriesWhitelist
      developerDisplayName
      description
      tags {
        id
        name
        groupName
      }
      pcReleaseDate
    }
  }
}
`

type GetCatalogOfferResponse struct {
	Data struct {
		Catalog struct {
			CatalogOffer *CatalogOffer `json:"catalogOffer"`
		} `json:"Catalog"`
	} `json:"data"`
}

type CatalogOffer struct {
	CountriesBlacklist []string `json:"countriesBlacklist"`
	CountriesWhitelist []string `json:"countriesWhitelist"`
	Tags               []OfferTag
}

type OfferTag struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	GroupName string `json:"groupName"`
}

func (catalogOfferResponse *CatalogOffer) GetTagsByGroupName(groupName string) []OfferTag {
	var result []OfferTag

	for _, tag := range catalogOfferResponse.Tags {
		if tag.GroupName == groupName {
			result = append(result, tag)
		}
	}

	return result
}

func GetCatalogOffer(sandboxId string, offerId string, locale string) (*CatalogOffer, error) {
	body := map[string]interface{}{
		"query": QUERY,
		"variables": map[string]string{
			"sandboxId": sandboxId,
			"offerId":   offerId,
			"locale":    locale,
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := getClient().
		Post(
			"https://graphql.epicgames.com/graphql",
			"application/json",
			bytes.NewReader(bodyBytes))

	if err != nil {
		return nil, err
	}

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response := &GetCatalogOfferResponse{}

	err = json.Unmarshal(respBodyBytes, response)
	if err != nil {
		return nil, err
	}

	if response.Data.Catalog.CatalogOffer == nil {
		return nil, errors.New("catalog offer not found")
	}

	return response.Data.Catalog.CatalogOffer, nil
}
