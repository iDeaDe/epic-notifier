package telegram

import (
	"bytes"
	"encoding/json"
)

type SendMessageRequest struct {
	ChatId              string    `json:"chat_id"`
	Text                string    `json:"text"`
	ParseMode           ParseMode `json:"parse_mode,omitempty"`
	DisableNotification bool      `json:"disable_notification,omitempty"`
	ProtectContent      bool      `json:"protect_content,omitempty"`
}

func (client *Client) SendMessage(request *SendMessageRequest) (*Message, error) {
	content, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	clientRequest := Request{
		Body:            bytes.NewBuffer(content),
		Name:            MethodSendMessage,
		postContentType: "application/json",
	}

	response, err := client.send(clientRequest)
	if err != nil {
		return nil, err
	}

	responseBody := Message{}

	return &responseBody, decodeResponse(response, &responseBody)
}
