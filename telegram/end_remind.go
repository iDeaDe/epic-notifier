package telegram

import (
	"encoding/json"
	"fmt"
	"github.com/ideade/epic-notifier/epicgames"
	"strconv"
)

type RemindPostResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageId int `json:"message_id"`
	} `json:"result"`
}

func (tg *Telegram) RemoveRemind(messageId int) error {
	getLogger().Println(fmt.Sprintf("Removing remind post(ID:%d)\n", messageId))
	queryParams := map[string]string{
		"chat_id":    tg.ChannelName,
		"message_id": strconv.Itoa(messageId),
	}
	req := Request{
		Method: MethodGet,
		Name:   "deleteMessage",
		Params: &queryParams,
		Body:   nil,
	}

	_, err := tg.Send(&req)

	if err != nil {
		return err
	}
	return nil
}

func (tg *Telegram) Remind(giveaway []epicgames.Game) (int, error) {
	getLogger().Println("Building keyboard buttons")
	keyboard := make(map[string][][1]KeyBoardButton)

	for _, game := range giveaway {
		gameButton := KeyBoardButton{
			Text: game.Title,
			Url:  game.Url,
		}

		keyboard["inline_keyboard"] = append(keyboard["inline_keyboard"], [1]KeyBoardButton{gameButton})
	}
	linkButton, err := json.Marshal(keyboard)
	if err != nil {
		return -1, err
	}

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

	getLogger().Println("Sending remind post to the Telegram API")
	resp, err := tg.Send(&req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	message := new(RemindPostResponse)
	err = json.NewDecoder(resp.Body).Decode(&message)
	if err != nil {
		return -1, err
	}

	return message.Result.MessageId, err
}
