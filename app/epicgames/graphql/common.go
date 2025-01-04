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

	requestUrl, _ := url.Parse("https://graphql.epicgames.com/graphql")
	requestUrl.RawQuery = query.Encode()

	resp, err := client.Get(requestUrl.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(response)
}
