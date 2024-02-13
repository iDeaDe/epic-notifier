package telegram

import (
	"io"
	"net/http"
	"net/url"
)

type ApiMethod = string

// API Methods
const (
	MethodSendMessage        ApiMethod = "sendMessage"
	MethodSendPhoto          ApiMethod = "sendPhoto"
	MethodSendMediaGroup     ApiMethod = "sendMediaGroup"
	MethodEditMessageCaption ApiMethod = "editMessageCaption"
	MethodDeleteMessage      ApiMethod = "deleteMessage"
)

type Request struct {
	Body   io.Reader
	Params *map[string]string
	Name   ApiMethod

	postContentType string
}

type Client struct {
	Token      string
	httpClient *http.Client
}

var Reserved = map[string]string{
	"<": "&lt;",
	">": "&gt;",
	"&": "&amp;",
}

var httpClient = http.DefaultClient

func NewClient(token string, httpClient *http.Client) *Client {
	return &Client{token, httpClient}
}

func (client *Client) send(request Request) (*http.Response, error) {
	// Собираем ссылку из параметров запроса
	reqUrl, err := url.Parse(ApiUrl)
	if err != nil {
		return nil, err
	}
	reqUrl.Path = "bot" + client.Token + "/" + request.Name
	qVal := reqUrl.Query()
	if request.Params != nil {
		for name, value := range *request.Params {
			qVal.Add(name, value)
		}
	}
	reqUrl.RawQuery = qVal.Encode()

	clientRequest, err := http.NewRequest(getHttpMethod(&request), reqUrl.String(), request.Body)
	if err != nil {
		return nil, err
	}

	if request.postContentType != "" {
		clientRequest.Header.Set("Content-Type", request.postContentType)
	}

	resp, err := httpClient.Do(clientRequest)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getHttpMethod(request *Request) string {
	if request.Body != nil {
		return http.MethodPost
	}

	return http.MethodGet
}