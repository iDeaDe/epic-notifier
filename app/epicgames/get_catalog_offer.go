package epicgames

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
	"net"
	"net/url"
	"strings"
	"time"
)

type CatalogOffer struct {
	DeveloperDisplayName string     `json:"developerDisplayName"`
	PublisherDisplayName string     `json:"publisherDisplayName"`
	Description          string     `json:"description"`
	CountriesBlacklist   []string   `json:"countriesBlacklist"`
	CountriesWhitelist   []string   `json:"countriesWhitelist"`
	Tags                 []OfferTag `json:"tags"`
	Price                price      `json:"price"`
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

const getCatalogOfferExtensions = "{\"persistedQuery\":{\"version\":1,\"sha256Hash\":\"abafd6e0aa80535c43676f533f0283c7f5214a59e9fae6ebfb37bed1b1bb2e9b\"}}"

func getCatalogOffer(locale string, country string, game *Game) (*CatalogOffer, error) {
	variables, _ := json.Marshal(map[string]any{
		"locale":    locale,
		"country":   country,
		"sandboxId": game.Namespace,
		"offerId":   game.Id,
	})

	graphqlUrl, _ := url.Parse("https://graphql.epicgames.com/graphql")
	urlBuf := strings.Builder{}

	params := map[string]string{
		"operationName": "getCatalogOffer",
		"variables":     string(variables),
		"extensions":    getCatalogOfferExtensions,
	}

	for name, value := range params {
		if urlBuf.Len() > 0 {
			urlBuf.WriteByte('&')
		}

		urlBuf.WriteString(name)
		urlBuf.WriteByte('=')
		urlBuf.WriteString(value)
	}

	graphqlUrl.RawQuery = urlBuf.String()

	addr, err := net.LookupHost(chromeHost)
	if err != nil {
		return nil, err
	}

	if len(addr) == 0 {
		return nil, errors.New("chrome host not found")
	}

	allocatorCtx, cancel := chromedp.NewRemoteAllocator(context.Background(), fmt.Sprintf("ws://%s:9222", addr[0]))
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	done := make(chan bool)
	var requestId network.RequestID

	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent:
			getLogger().Debug(
				"EventRequestWillBeSent",
				zap.String("requestId", string(ev.RequestID)),
				zap.String("requestUrl", ev.Request.URL))
			if ev.Request.URL == graphqlUrl.String() {
				requestId = ev.RequestID
			}
		case *network.EventLoadingFinished:
			getLogger().Debug("EventLoadingFinished", zap.String("requestId", string(ev.RequestID)))
			if ev.RequestID == requestId {
				close(done)
			}
		}
	})

	if err = chromedp.Run(ctx,
		chromedp.Navigate(graphqlUrl.String()),
	); err != nil {
		return nil, err
	}

	<-done

	var respBodyBytes []byte

	if err = chromedp.Run(
		ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error

			respBodyBytes, err = network.GetResponseBody(requestId).Do(ctx)

			return err
		}),
	); err != nil {
		return nil, err
	}

	getLogger().Debug(string(respBodyBytes))

	response := &getCatalogOfferResponse{}

	err = json.Unmarshal(respBodyBytes, response)
	if err != nil {
		return nil, err
	}

	if response.Data.Catalog.CatalogOffer == nil {
		return nil, errors.New("catalog offer not found")
	}

	return response.Data.Catalog.CatalogOffer, nil
}
