package telegram

type DeleteMessageRequest struct {
	ChatId    string
	MessageId string
}

func (client *Client) DeleteMessage(request *DeleteMessageRequest) (bool, error) {
	clientRequest := Request{
		Params: &map[string]string{
			"chat_id":    request.ChatId,
			"message_id": request.MessageId,
		},
		Name: MethodDeleteMessage,
	}

	response, err := client.send(clientRequest)
	if err != nil {
		return false, err
	}

	result := false

	return result, decodeResponse(response, &result)
}
