package telegram

import (
	"encoding/json"
	"fmt"
	"github.com/ideade/epic-notifier/epicgames"
	"strconv"
	"strings"
	"time"
)

type InputMedia struct {
	Type      string `json:"type"`
	Url       string `json:"media"`
	Caption   string `json:"caption"`
	ParseMode string `json:"parse_mode"`
}

type NewPostResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageId int `json:"message_id"`
	} `json:"result"`
}

func formatPostText(games *[]epicgames.Game, nextGiveawayTime time.Time) (string, error) {
	var gameTitles []string

	for index, game := range *games {
		EscapeString(&game.Title)
		gameTitle := fmt.Sprintf("%d. <b><a href=\"%s\">%s</a></b>", index+1, game.Url, game.Title)

		gameTitles = append(gameTitles, gameTitle)
	}

	moscowLoc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return "", err
	}

	messageText := fmt.Sprintf(
		"Анонс следующей раздачи:\n%s\n\nДата начала раздачи: %s",
		strings.Join(gameTitles, "\n"),
		fmt.Sprintf("%d %s, %s",
			nextGiveawayTime.Day(),
			epicgames.GetMonth(nextGiveawayTime.Month()),
			nextGiveawayTime.In(moscowLoc).Format("15:04 MST")))

	return messageText, nil
}

func (tg *Telegram) RemoveNextPost(messageId int) error {
	getLogger().Println(fmt.Sprintf("Removing old next giveaway post(ID:%d)", messageId))
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

	getLogger().Println("Sending request to the Telegram API, request URL")
	_, err := tg.Send(&req)

	if err != nil {
		return err
	}
	return nil
}

func (tg *Telegram) UpdateNext(messageId int, ga *epicgames.Giveaway) error {
	getLogger().Println(fmt.Sprintf("Updating next giveaway post(ID:%d)", messageId))

	formattedPostText, err := formatPostText(&ga.NextGames, ga.Next)
	if err != nil {
		return err
	}

	queryParams := map[string]string{
		"chat_id":    tg.ChannelName,
		"message_id": strconv.Itoa(messageId),
		"caption":    formattedPostText,
		"parse_mode": "HTML",
	}
	req := Request{
		Method: MethodGet,
		Name:   "editMessageCaption",
		Params: &queryParams,
		Body:   nil,
	}

	getLogger().Println("Sending request to the Telegram API")
	resp, err := tg.Send(&req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 {
		getLogger().Println("Message text didn't changed")
	} else {
		getLogger().Println(resp.Status)
	}

	return nil
}

func (tg *Telegram) PostNext(ga *epicgames.Giveaway) (int, error) {
	var media []InputMedia

	for _, game := range ga.NextGames {
		if game.Image == "" {
			continue
		}

		photo := InputMedia{
			Type: "photo",
			Url:  game.Image,
		}

		media = append(media, photo)
	}

	if len(media) > 0 {
		formattedPostText, err := formatPostText(&ga.NextGames, ga.Next)
		if err != nil {
			return -1, err
		}
		media[0].Caption = formattedPostText
		media[0].ParseMode = "HTML"
	} else {
		return -1, nil
	}

	jsonMedia, err := json.Marshal(media)
	if err != nil {
		return -1, err
	}

	queryParams := map[string]string{
		"chat_id":              tg.ChannelName,
		"media":                string(jsonMedia),
		"disable_notification": "True",
	}
	req := Request{
		Method: MethodGet,
		Name:   "sendMediaGroup",
		Params: &queryParams,
		Body:   nil,
	}

	getLogger().Println("Sending request to the Telegram API")
	resp, err := tg.Send(&req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	message := new(NewPostResponse)
	err = json.NewDecoder(resp.Body).Decode(&message)
	if err != nil {
		return -1, err
	}

	return message.Result.MessageId, nil
}
