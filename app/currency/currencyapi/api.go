package currencyapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/ideade/epic-notifier/app/currency"
)

const baseUrl = "https://api.currencyapi.com/v3"

type exchangeDataResponse struct {
	Data map[string]struct {
		Code  string  `json:"code"`
		Value float64 `json:"value"`
	} `json:"data"`
}

type apiClient struct {
	httpClient *http.Client
	token      string
}

func newApiClient(httpClient *http.Client, token string) *apiClient {
	return &apiClient{httpClient, token}
}

func (client *apiClient) getRates(baseCurrency string, currencies []string) (map[currency.Pair]float64, error) {
	query := url.Values{}
	query.Set("apikey", client.token)
	query.Set("base_currency", baseCurrency)
	for _, currencyItem := range currencies {
		query.Add("currencies[]", currencyItem)
	}

	requestUrl, _ := url.Parse(baseUrl)
	requestUrl = requestUrl.JoinPath("latest")
	requestUrl.RawQuery = query.Encode()

	response, err := client.httpClient.Get(requestUrl.String())
	if err != nil {
		return nil, err
	}

	bodyBytes := bytes.Buffer{}
	_, err = bodyBytes.ReadFrom(response.Body)
	if err != nil {
		return nil, err
	}

	responseStruct := exchangeDataResponse{}

	err = json.Unmarshal(bodyBytes.Bytes(), &responseStruct)
	if err != nil {
		return nil, err
	}

	rates := make(map[currency.Pair]float64)

	for _, rate := range responseStruct.Data {
		pair := currency.NewPair(baseCurrency, rate.Code)
		rates[*pair] = rate.Value
	}

	return rates, nil
}
