package telegram

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const ApiUrl = "https://api.telegram.org/"

const (
	MethodGet = iota
	MethodPost
)

type Telegram struct {
	Token       string
	ChannelName string
}

type Request struct {
	Method uint
	Name   string
	Params *map[string]string
	Body   io.Reader
}

var Reserved = map[string]string{
	"<": "&lt;",
	">": "&gt;",
	"&": "&amp;",
}

var logger *log.Logger

func getLogger() *log.Logger {
	if logger == nil {
		logger = log.Default()
	}

	return logger
}

func SetLogger(newLogger *log.Logger) {
	logger = newLogger
}

func EscapeString(text *string) {
	for search, replace := range Reserved {
		*text = strings.ReplaceAll(*text, search, replace)
	}
}

func (tg *Telegram) Send(req *Request) (*http.Response, error) {
	// Собираем ссылку из параметров запроса
	reqUrl, err := url.Parse(ApiUrl)
	if err != nil {
		return nil, err
	}
	reqUrl.Path = "bot" + tg.Token + "/" + req.Name
	qVal := reqUrl.Query()
	if req.Params != nil {
		for name, value := range *req.Params {
			qVal.Add(name, value)
		}
	}
	reqUrl.RawQuery = qVal.Encode()

	var resp *http.Response
	switch req.Method {
	case MethodGet:
		resp, err = http.Get(reqUrl.String())
		break
	case MethodPost:
		resp, err = http.Post(reqUrl.String(), "application/json", req.Body)
	default:
		return nil, errors.New("unknown request method")
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// JoinNotEmptyStrings Not recommended to use in your projects
func JoinNotEmptyStrings(elems []string, sep string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return elems[0]
	}

	result := elems[0]

	for _, item := range elems[1:] {
		if item != "" {
			result += sep + item
		}
	}

	return result
}
