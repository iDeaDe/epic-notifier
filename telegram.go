package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const ApiUrl = "https://api.telegram.org/"

type TelegramSettings struct {
	Token       string
	ChannelName string
}

type KeyBoardButton struct {
	Text string `json:"text"`
	Url  string `json:"url"`
}

var Reserved = map[string]string{
	"<": "&lt;",
	">": "&gt;",
	"&": "&amp;",
}

func EscapeString(text *string) {
	for search, replace := range Reserved {
		*text = strings.ReplaceAll(*text, search, replace)
	}
}

func (tg *TelegramSettings) Post(game *Game, silent bool) {
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

	reqUrl, _ := url.Parse(ApiUrl)
	reqUrl.Path = "bot" + tg.Token + "/sendPhoto"

	// Параметры для запроса к Telegram API
	qVal := reqUrl.Query()
	qVal.Add("chat_id", tg.ChannelName)
	qVal.Add("photo", game.Image)
	qVal.Add("parse_mode", "HTML")
	qVal.Add("reply_markup", string(linkButton))
	if silent {
		qVal.Add("disable_notification", "True")
	}

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

	messageText := fmt.Sprintf(
		"<b>Раздаётся игра %s</b>\n\n%s%s\nИгра доступна бесплатно до %s",
		game.Title,
		publisher,
		developer,
		fmt.Sprintf("%d %s",
			game.Date.End.Day(),
			GetMonth(game.Date.End.Month())))

	qVal.Add("caption", messageText)

	reqUrl.RawQuery = qVal.Encode()

	log.Println("Sending request to the Telegram API, request URL")
	resp, err := http.Get(reqUrl.String())
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			log.Println(resp.StatusCode)
		}
		log.Fatal(err)
	}
}
