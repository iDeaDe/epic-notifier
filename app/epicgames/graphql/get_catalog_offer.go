package graphql

type CatalogOffer struct {
	DeveloperDisplayName string     `json:"developerDisplayName"`
	PublisherDisplayName string     `json:"publisherDisplayName"`
	Description          string     `json:"description"`
	CountriesBlacklist   []string   `json:"countriesBlacklist"`
	CountriesWhitelist   []string   `json:"countriesWhitelist"`
	Tags                 []OfferTag `json:"tags"`
	Price                struct {
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
}

type OfferTag struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	GroupName string `json:"groupName"`
}

type getCatalogOfferResponse struct {
	Data struct {
		Catalog struct {
			CatalogOffer *CatalogOffer `json:"catalogOffer"`
		} `json:"Catalog"`
	} `json:"data"`
	Extensions struct {
	} `json:"extensions"`
}

const getCatalogOfferExtensions extensions = "{\"persistedQuery\":{\"version\":1,\"sha256Hash\":\"abafd6e0aa80535c43676f533f0283c7f5214a59e9fae6ebfb37bed1b1bb2e9b\"}}"

func GetCatalogOffer(locale, country, offerId, sandboxId string) (*CatalogOffer, error) {
	response := getCatalogOfferResponse{}

	err := doRequest(
		graphqlRequest{
			operationName: "getCatalogOffer",
			variables: map[string]any{
				"locale":    locale,
				"country":   country,
				"sandboxId": sandboxId,
				"offerId":   offerId,
			},
			extensions: getCatalogOfferExtensions,
		},
		&response,
	)
	if err != nil {
		return nil, err
	}

	return response.Data.Catalog.CatalogOffer, nil
}
