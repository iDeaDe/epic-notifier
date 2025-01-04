package currency

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
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

func (client *apiClient) getRates(baseCurrency string, currencies []string) (map[Pair]float64, error) {
	requestUrl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	requestUrl = requestUrl.JoinPath("latest")
	qVal := requestUrl.Query()
	qVal.Add("apikey", client.token)
	qVal.Add("base_currency", baseCurrency)
	for _, currency := range currencies {
		qVal.Add("currencies[]", currency)
	}
	requestUrl.RawQuery = qVal.Encode()

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

	rates := make(map[Pair]float64)

	for _, rate := range responseStruct.Data {
		pair := NewPair(baseCurrency, rate.Code)
		rates[*pair] = rate.Value
	}

	return rates, nil
}
