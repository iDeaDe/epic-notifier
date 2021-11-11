package telegram

import (
	"encoding/json"
	"fmt"
	"github.com/ideade/epic-notifier/epicgames"
	"log"
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
	Result []struct {
		MessageId int `json:"message_id"`
	} `json:"result"`
}

func formatPostText(games *[]epicgames.Game, nextGiveawayTime time.Time) string {
	var gameTitles []string

	for index, game := range *games {
		EscapeString(&game.Title)
		gameTitle := fmt.Sprintf("%d. <b><a href=\"%s\">%s</a></b>", index+1, game.Url, game.Title)

		gameTitles = append(gameTitles, gameTitle)
	}

	moscowLoc, _ := time.LoadLocation("Europe/Moscow")

	messageText := fmt.Sprintf(
		"Анонс следующей раздачи:\n%s\n\nДата начала раздачи: %s",
		strings.Join(gameTitles, "\n"),
		fmt.Sprintf("%d %s, %s",
			nextGiveawayTime.Day(),
			epicgames.GetMonth(nextGiveawayTime.Month()),
			nextGiveawayTime.In(moscowLoc).Format("15:04 MST")))

	return messageText
}

func (tg *Settings) RemoveNextPost(messageId string) error {
	log.Printf("Removing old next giveaway post(ID:%s)\n", messageId)
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

func (tg *Settings) UpdateNext(messageId string, ga *epicgames.Giveaway) {
	log.Printf("Updating next giveaway post(ID:%s)\n", messageId)
	queryParams := map[string]string{
		"chat_id":    tg.ChannelName,
		"message_id": messageId,
		"caption":    formatPostText(&ga.NextGames, ga.Next),
		"parse_mode": "HTML",
	}
	req := Request{
		Method: MethodGet,
		Name:   "editMessageCaption",
		Params: &queryParams,
		Body:   nil,
	}

	log.Println("Sending request to the Telegram API")
	resp, err := tg.Send(&req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 {
		log.Println("Message text didn't changed")
	} else {
		log.Println(resp.Status)
	}
}

func (tg *Settings) PostNext(ga *epicgames.Giveaway) int {
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
		media[0].Caption = formatPostText(&ga.NextGames, ga.Next)
		media[0].ParseMode = "HTML"
	} else {
		return -1
	}

	jsonMedia, _ := json.Marshal(media)

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

	log.Println("Sending request to the Telegram API")
	resp, err := tg.Send(&req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	message := new(NewPostResponse)
	_ = json.NewDecoder(resp.Body).Decode(&message)

	return message.Result[0].MessageId
}
