package exchangerate

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ideade/epic-notifier/app/currency"
)

type ExchangeRates struct {
	Result             string
	TimeNextUpdateUnix int                `json:"time_next_update_unix"`
	ConversionRates    map[string]float64 `json:"conversion_rates"`
}

type apiClient struct {
	httpClient  *http.Client
	preparedUrl string
}

func newApiClient(httpClient *http.Client, token string) *apiClient {
	return &apiClient{
		httpClient:  httpClient,
		preparedUrl: strings.Replace(serviceUrl, "{{token}}", token, -1),
	}
}

func (client *apiClient) getRates(baseCurrency string, currencies []string) (map[currency.Pair]float64, error) {
	url := strings.Replace(client.preparedUrl, "{{currency}}", baseCurrency, -1)

	resp, err := client.httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	ratesResponse := ExchangeRates{}

	err = json.NewDecoder(resp.Body).Decode(&ratesResponse)
	if err != nil {
		return nil, err
	}

	rates := make(map[currency.Pair]float64)

	for _, targetCurrency := range currencies {
		if val, ok := ratesResponse.ConversionRates[targetCurrency]; ok {
			pair := currency.NewPair(baseCurrency, targetCurrency)
			rates[*pair] = val
		}
	}

	return rates, nil
}
