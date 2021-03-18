package telegram

import (
	"encoding/json"
	"log"
)

type RemindPostResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageId int `json:"message_id"`
	} `json:"result"`
}

func (tg *TelegramSettings) RemoveRemind(messageId string) error {
	log.Printf("Removing remind post(ID:%s)\n", messageId)
	queryParams := map[string]string{
		"chat_id":    tg.ChannelName,
		"message_id": messageId,
	}
	req := Request{
		Method: MethodGet,
		Name:   "deleteMessage",
		Params: &queryParams,
		Body:   nil,
	}

	log.Println("Sending request to the Telegram API, request URL")
	_, err := tg.Send(&req)

	if err != nil {
		return err
	}
	return nil
}

func (tg *TelegramSettings) Remind() int {
	queryParams := map[string]string{
		"chat_id":    tg.ChannelName,
		"text":       "<b>Напоминание!</b>\n\nСегодня последний день раздачи.",
		"parse_mode": "HTML",
	}
	req := Request{
		Method: MethodGet,
		Name:   "sendMessage",
		Params: &queryParams,
		Body:   nil,
	}

	log.Println("Sending remind post to the Telegram API")
	resp, err := tg.Send(&req)

	if err != nil {
		log.Fatal(err)
	}

	message := new(RemindPostResponse)
	_ = json.NewDecoder(resp.Body).Decode(&message)

	return message.Result.MessageId
}
