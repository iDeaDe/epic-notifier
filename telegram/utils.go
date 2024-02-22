package telegram

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
)

const ApiUrl = "https://api.telegram.org/"

type Telegram struct {
	Token       string
	ChannelName string
}

type ParseMode string

type Response struct {
	Result      interface{} `json:"result"`
	Description string      `json:"description"`
	Ok          bool        `json:"ok"`
}

var (
	ParseModeHtml     ParseMode = "HTML"
	ParseModeMarkdown ParseMode = "MarkdownV2"
)

var (
	UnknownMethodErr  = errors.New("unknown request method")
	DefaultRequestErr = errors.New("the request failed")
	EmptyBodyErr      = errors.New("response body is empty")
)

func decodeResponse(response *http.Response, target any) error {
	bodyStruct := Response{}
	bodyStruct.Result = target

	bodyContent, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if len(bodyContent) == 0 {
		return EmptyBodyErr
	}

	if err = json.Unmarshal(bodyContent, &bodyStruct); err != nil {
		return err
	}

	if !bodyStruct.Ok {
		if bodyStruct.Description != "" {
			return errors.New(bodyStruct.Description)
		} else {
			return DefaultRequestErr
		}
	}

	return nil
}

func addParamsToForm(form *multipart.Writer, params map[string]string) error {
	for name, value := range params {
		if err := form.WriteField(name, value); err != nil {
			return err
		}
	}

	return nil
}
