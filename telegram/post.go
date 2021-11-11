package telegram

import (
	"encoding/json"
	"fmt"
	"github.com/ideade/epic-notifier/epicgames"
	"log"
	"strings"
	"time"
)

type KeyBoardButton struct {
	Text string `json:"text"`
	Url  string `json:"url"`
}

func (tg *Settings) Post(game *epicgames.Game, silent bool) {
	log.Println("Building keyboard buttons")

	// Интерактивные кнопки внизу поста
	var shopButton [1]KeyBoardButton
	shopButton[0] = KeyBoardButton{
		Text: "Страница игры",
		Url:  game.Url,
	}
	var moreGamesButton [1]KeyBoardButton
	moreGamesButton[0] = KeyBoardButton{
		Text: "Другие бесплатные игры",
		Url:  "https://www.epicgames.com/store/ru/free-games",
	}
	keyboard := make(map[string][2][1]KeyBoardButton)
	keyboard["inline_keyboard"] = [2][1]KeyBoardButton{shopButton, moreGamesButton}
	linkButton, _ := json.Marshal(keyboard)

	publisher := ""
	developer := ""

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

	description := ""
	if len(game.Description) > 10 {
		description = fmt.Sprintf("\n%s\n", game.Description)
	}

	price := ""
	if game.Price.Format != "" {
		price = fmt.Sprintf("Обычная цена: <b>%s</b>", game.Price.Format)
	}

	moscowLoc, _ := time.LoadLocation("Europe/Moscow")

	endDate := fmt.Sprintf(
		"\nИгра доступна бесплатно до %d %s, %s",
		game.Date.End.Day(),
		epicgames.GetMonth(game.Date.End.Month()),
		game.Date.End.In(moscowLoc).Format("15:04 MST"))

	messageText := strings.Join(
		[]string{title, description, price, publisher, developer, endDate},
		"\n")

	queryParams := map[string]string{
		"chat_id":      tg.ChannelName,
		"photo":        game.Image,
		"parse_mode":   "HTML",
		"reply_markup": string(linkButton),
		"caption":      messageText,
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

	log.Println("Sending request to the Telegram API, request URL")
	resp, err := tg.Send(&req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	message := new(NewPostResponse)
	_ = json.NewDecoder(resp.Body).Decode(&message)
}
