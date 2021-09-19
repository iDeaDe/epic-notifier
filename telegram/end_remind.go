package telegram

import (
	"encoding/json"
	"github.com/ideade/epic-notifier/epicgames"
	"log"
)

type RemindPostResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageId int `json:"message_id"`
	} `json:"result"`
}

func (tg *Settings) RemoveRemind(messageId string) error {
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

func (tg *Settings) Remind(giveaway []epicgames.Game) int {
	log.Println("Building keyboard buttons")
	keyboard := make(map[string][][1]KeyBoardButton)

	for _, game := range giveaway {
		gameButton := KeyBoardButton{
			Text: game.Title,
			Url:  game.Url,
		}

		keyboard["inline_keyboard"] = append(keyboard["inline_keyboard"], [1]KeyBoardButton{gameButton})
	}
	linkButton, _ := json.Marshal(keyboard)

	queryParams := map[string]string{
		"chat_id":      tg.ChannelName,
		"text":         "<b>Напоминание!</b>\n\nСегодня последний день раздачи.",
		"parse_mode":   "HTML",
		"reply_markup": string(linkButton),
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
	defer resp.Body.Close()

	message := new(RemindPostResponse)
	_ = json.NewDecoder(resp.Body).Decode(&message)

	return message.Result.MessageId
}
