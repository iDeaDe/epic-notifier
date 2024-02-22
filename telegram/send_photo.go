package telegram

import (
	"bytes"
	"io"
	"mime/multipart"
	"strconv"
)

type SendPhotoRequest struct {
	Photo               io.ReadCloser
	ChatId              string
	Caption             string
	ParseMode           ParseMode
	HasSpoiler          bool
	DisableNotification bool
	ProtectContent      bool
}

func (client *Client) SendPhoto(request *SendPhotoRequest) (*Message, error) {
	body := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(body)
	imageWriter, err := bodyWriter.CreateFormFile("photo", "label")
	_, err = io.Copy(imageWriter, request.Photo)
	if err != nil {
		return nil, err
	}

	params := map[string]string{
		"chat_id":              request.ChatId,
		"caption":              request.Caption,
		"parse_mode":           string(request.ParseMode),
		"has_spoiler":          strconv.FormatBool(request.HasSpoiler),
		"disable_notification": strconv.FormatBool(request.DisableNotification),
		"protect_content":      strconv.FormatBool(request.ProtectContent),
	}

	if err = addParamsToForm(bodyWriter, params); err != nil {
		return nil, err
	}

	if err = request.Photo.Close(); err != nil {
		return nil, err
	}

	if err = bodyWriter.Close(); err != nil {
		return nil, err
	}

	clientRequest := Request{
		Body: body,
		Name: MethodSendPhoto,
	}

	clientRequest.postContentType = bodyWriter.FormDataContentType()

	response, err := client.send(clientRequest)
	if err != nil {
		return nil, err
	}

	responseBody := Message{}

	return &responseBody, decodeResponse(response, &responseBody)
}
