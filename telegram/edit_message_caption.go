package telegram

type EditMessageCaptionRequest struct {
	ChatId    string
	MessageId string
	Caption   string
	ParseMode ParseMode
}

func (client *Client) EditMessageCaption(request *EditMessageCaptionRequest) (*Message, error) {
	clientRequest := Request{
		Params: &map[string]string{
			"chat_id":    request.ChatId,
			"message_id": request.MessageId,
			"caption":    request.Caption,
			"parse_mode": string(request.ParseMode),
		},
		Name: MethodEditMessageCaption,
	}

	response, err := client.send(clientRequest)
	if err != nil {
		return nil, err
	}

	message := Message{}

	return &message, decodeResponse(response, &message)
}
