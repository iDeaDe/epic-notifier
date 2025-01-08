package currency

import (
	"net/http"
	"sync"
)

type Updater interface {
	SetErrorHandler(func(error))
	AddPair(Pair)
	Convert(float64, Pair) float64
	Update()
}

type BackgroundUpdater struct {
	apiClient    *apiClient
	errorHandler func(error)

	mu         sync.RWMutex
	rates      map[Pair]float64
	currencies map[string][]string
}

func NewBackgroundUpdater(client *http.Client, token string) *BackgroundUpdater {
	return &BackgroundUpdater{
		apiClient:  newApiClient(client, token),
		rates:      map[Pair]float64{},
		currencies: map[string][]string{},
	}
}

func (backgroundUpdater *BackgroundUpdater) SetErrorHandler(handler func(error)) {
	backgroundUpdater.errorHandler = handler
}

func (backgroundUpdater *BackgroundUpdater) AddPair(pair Pair) {
	backgroundUpdater.mu.Lock()
	defer backgroundUpdater.mu.Unlock()

	backgroundUpdater.currencies[pair.From] = append(backgroundUpdater.currencies[pair.From], pair.To)
}

func (backgroundUpdater *BackgroundUpdater) Convert(sum float64, pair Pair) float64 {
	backgroundUpdater.mu.RLock()
	defer backgroundUpdater.mu.RUnlock()

	if rate, ok := backgroundUpdater.rates[pair]; ok {
		return sum * rate
	}

	return sum
}

func (backgroundUpdater *BackgroundUpdater) Update() {
	go func() {
		newRates := map[Pair]float64{}

		for baseCurrency, currencies := range backgroundUpdater.currencies {
			rates, err := backgroundUpdater.apiClient.getRates(baseCurrency, currencies)
			if err != nil && backgroundUpdater.errorHandler != nil {
				backgroundUpdater.errorHandler(err)
			}

			for pair, rate := range rates {
				newRates[pair] = rate
			}
		}

		backgroundUpdater.mu.Lock()
		backgroundUpdater.rates = newRates
		backgroundUpdater.mu.Unlock()
	}()
}
