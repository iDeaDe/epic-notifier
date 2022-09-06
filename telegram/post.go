package telegram

import (
	"encoding/json"
	"fmt"
	"github.com/ideade/epic-notifier/epicgames"
	"log"
	"time"
)

type KeyBoardButton struct {
	Text string `json:"text"`
	Url  string `json:"url"`
}

func (tg *Settings) Post(game *epicgames.Game, silent bool) {
	description := ""
	publisher := ""
	developer := ""
	price := ""

	EscapeString(&game.Title)
	EscapeString(&game.Publisher)
	EscapeString(&game.Developer)
	EscapeString(&game.Price.Format)

	title := fmt.Sprintf("<b>Раздаётся игра %s</b>", game.Title)

	if game.Publisher != "" {
		publisher = fmt.Sprintf("Издатель: <b>%s</b>", game.Publisher)
	}

	if game.Developer != "" {
		developer = fmt.Sprintf("Разработчик: <b>%s</b>", game.Developer)
	}

	if game.Description != "" {
		description = fmt.Sprintf("\n%s\n", game.Description)
	}

	if game.Price.Format != "" && game.Price.Original > 0 {
		price = fmt.Sprintf("Обычная цена: <b>%s</b>", game.Price.Format)
	}

	moscowLoc, err := time.LoadLocation("Europe/Moscow")

	if err != nil {
		log.Panicln(err)
	}

	endDate := fmt.Sprintf(
		"\nИгра доступна бесплатно до %d %s, %s",
		game.Date.End.Day(),
		epicgames.GetMonth(game.Date.End.Month()),
		game.Date.End.In(moscowLoc).Format("15:04 MST"))

	messageText := JoinNotEmptyStrings(
		[]string{
			title,
			description,
			price,
			publisher,
			developer,
			fmt.Sprintf("\n<a href=\"%s\">Страница в магазине</a>", game.Url),
			endDate,
		},
		"\n")

	queryParams := map[string]string{
		"chat_id":    tg.ChannelName,
		"photo":      game.Image,
		"parse_mode": "HTML",
		"caption":    messageText,
	}
	if silent {
		queryParams["disable_notification"] = "True"
	}
	req := Request{
		Method: MethodGet,
		Name:   "sendPhoto",
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
}
