package currency

import (
	"maps"
	"net/http"
)

var exchangeRates = map[Pair]float64{}

type Updater struct {
	currencies        map[string][]string
	currencyApiClient *apiClient
}

func NewUpdater(client *http.Client, token string) *Updater {
	return &Updater{currencyApiClient: newApiClient(client, token), currencies: map[string][]string{}}
}

func (updater *Updater) AddPair(pair Pair) {
	updater.currencies[pair.From] = append(updater.currencies[pair.From], pair.To)
}

func (updater *Updater) Update() error {
	for baseCurrency, currencies := range updater.currencies {
		rates, err := updater.currencyApiClient.getRates(baseCurrency, currencies)
		if err != nil {
			return err
		}

		maps.Copy(exchangeRates, rates)
	}

	return nil
}
