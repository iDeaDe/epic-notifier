package graphql

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type extensions string

type graphqlRequest struct {
	operationName string
	variables     map[string]any
	extensions    extensions
}

var client *http.Client

func SetClient(newClient *http.Client) {
	client = newClient
}

func doRequest(req graphqlRequest, response any) error {
	variables, err := json.Marshal(req.variables)
	if err != nil {
		return err
	}

	query := url.Values{}
	query.Set("operationName", req.operationName)
	query.Set("variables", string(variables))
	query.Set("extensions", string(req.extensions))

	requestUrl, _ := url.Parse("https://store.epicgames.com/graphql")
	requestUrl.RawQuery = query.Encode()

	request, _ := http.NewRequest(http.MethodGet, requestUrl.String(), nil)
	request.Header.Set("User-Agent", "okhttp/5.1.0")

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(response)
}
