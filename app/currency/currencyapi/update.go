package currencyapi

import (
	"net/http"
	"sync"

	"github.com/ideade/epic-notifier/app/currency"
)

type BackgroundUpdater struct {
	apiClient    *apiClient
	errorHandler func(error)

	mu         sync.RWMutex
	rates      map[currency.Pair]float64
	currencies map[string][]string
}

func NewBackgroundUpdater(client *http.Client, token string) *BackgroundUpdater {
	return &BackgroundUpdater{
		apiClient:  newApiClient(client, token),
		rates:      map[currency.Pair]float64{},
		currencies: map[string][]string{},
	}
}

func (updater *BackgroundUpdater) SetErrorHandler(handler func(error)) {
	updater.errorHandler = handler
}

func (updater *BackgroundUpdater) AddPair(pair currency.Pair) {
	updater.mu.Lock()
	defer updater.mu.Unlock()

	updater.currencies[pair.From] = append(updater.currencies[pair.From], pair.To)
}

func (updater *BackgroundUpdater) Convert(sum float64, pair currency.Pair) float64 {
	updater.mu.RLock()
	defer updater.mu.RUnlock()

	if rate, ok := updater.rates[pair]; ok {
		return sum * rate
	}

	return sum
}

func (updater *BackgroundUpdater) Update() {
	go func() {
		newRates := map[currency.Pair]float64{}

		for baseCurrency, currencies := range updater.currencies {
			rates, err := updater.apiClient.getRates(baseCurrency, currencies)
			if err != nil && updater.errorHandler != nil {
				updater.errorHandler(err)
			}

			for pair, rate := range rates {
				newRates[pair] = rate
			}
		}

		updater.mu.Lock()
		updater.rates = newRates
		updater.mu.Unlock()
	}()
}
