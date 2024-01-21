package telegram

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"strconv"
)

type InputMedia interface {
	GetType() string
}

type InputMediaPhoto struct {
	Media      io.ReadCloser
	Caption    string
	ParseMode  ParseMode
	HasSpoiler bool
}

func (photo *InputMediaPhoto) GetType() string {
	return "photo"
}

type SendMediaGroupRequest struct {
	Media               []InputMediaPhoto
	ChatId              string
	DisableNotification bool
	ProtectContent      bool
}

func (client *Client) SendMediaGroup(request *SendMediaGroupRequest) ([]Message, error) {
	body := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(body)

	var media []map[string]any

	for index, mediaInput := range request.Media {
		filename := "file" + strconv.Itoa(index)
		fileWriter, err := bodyWriter.CreateFormFile(filename, filename)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(fileWriter, mediaInput.Media)
		if err != nil {
			return nil, err
		}

		if err = mediaInput.Media.Close(); err != nil {
			return nil, err
		}

		media = append(media, map[string]any{
			"type":        mediaInput.GetType(),
			"media":       "attach://" + filename,
			"caption":     mediaInput.Caption,
			"parse_mode":  string(mediaInput.ParseMode),
			"has_spoiler": mediaInput.HasSpoiler,
		})
	}

	mediaParamMarshalled, err := json.Marshal(media)
	if err != nil {
		return nil, err
	}

	params := map[string]string{
		"chat_id":              request.ChatId,
		"disable_notification": strconv.FormatBool(request.DisableNotification),
		"protect_content":      strconv.FormatBool(request.ProtectContent),
		"media":                string(mediaParamMarshalled),
	}

	if err = addParamsToForm(bodyWriter, params); err != nil {
		return nil, err
	}

	if err = bodyWriter.Close(); err != nil {
		return nil, err
	}

	clientRequest := Request{
		Body: body,
		Name: MethodSendMediaGroup,
	}

	clientRequest.postContentType = bodyWriter.FormDataContentType()

	response, err := client.send(clientRequest)
	if err != nil {
		return nil, err
	}

	var responseBody []Message

	return responseBody, decodeResponse(response, &responseBody)
}
