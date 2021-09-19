package telegram

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const ApiUrl = "https://api.telegram.org/"

const (
	MethodGet = iota
	MethodPost
)

type Settings struct {
	Token            string
	ChannelName      string
	ChannelModerator string
	Webhook          string
}

type Request struct {
	Method uint
	Name   string
	Params *map[string]string
	Body   *io.Reader
}

var Reserved = map[string]string{
	"<": "&lt;",
	">": "&gt;",
	"&": "&amp;",
}

func EscapeString(text *string) {
	for search, replace := range Reserved {
		*text = strings.ReplaceAll(*text, search, replace)
	}
}

func (tg *Settings) Send(req *Request) (*http.Response, error) {
	// Собираем ссылку из параметров запроса
	reqUrl, _ := url.Parse(ApiUrl)
	reqUrl.Path = "bot" + tg.Token + "/" + req.Name
	qVal := reqUrl.Query()
	if req.Params != nil {
		for name, value := range *req.Params {
			qVal.Add(name, value)
		}
	}
	reqUrl.RawQuery = qVal.Encode()

	var resp *http.Response
	var err error
	// Отправляем запрос в зависимости от выбранного метода
	switch req.Method {
	case MethodGet:
		resp, err = http.Get(reqUrl.String())
		break
	case MethodPost:
		resp, err = http.Post(reqUrl.String(), "application/json", *req.Body)
	default:
		return nil, errors.New("unknown request method")
	}

	// Проверяем ошибки. В данном случае нас интересуют ошибки самого запроса, либо ошибки, связанные конкретно с Telegram API
	if err != nil {
		return nil, err
	}

	return resp, nil
}
