package telegram

import (
	"encoding/json"
	"fmt"
	"github.com/ideade/epic-notifier/epicgames"
	"log"
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
		Text: "Больше бесплатных игр",
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

	if game.Publisher != "" {
		publisher = fmt.Sprintf("Издатель: <b>%s</b>\n", game.Publisher)
	}

	if game.Developer != "" {
		developer = fmt.Sprintf("Разработчик: <b>%s</b>\n", game.Developer)
	}

	description := ""
	if len(game.Description) > 10 {
		description = fmt.Sprintf("\n%s\n", game.Description)
	}

	messageText := fmt.Sprintf(
		"<b>Раздаётся игра %s</b>%s\n\n%s%s\nИгра доступна бесплатно до %s",
		game.Title,
		description,
		publisher,
		developer,
		fmt.Sprintf("%d %s",
			game.Date.End.Day(),
			epicgames.GetMonth(game.Date.End.Month()),
		),
	)

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
