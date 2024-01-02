package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ideade/epic-notifier/epicgames"
	"html/template"
	"path/filepath"
	"time"
)

type KeyBoardButton struct {
	Text string `json:"text"`
	Url  string `json:"url"`
}

func (tg *Telegram) SendGame(game *epicgames.Game) error {
	tpl, err := template.New("game.gohtml").
		Funcs(
			template.FuncMap{
				"month": epicgames.GetMonth,
			},
		).
		ParseFiles(filepath.Join("template", "game.gohtml"))
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err = tpl.Execute(buf, game); err != nil {
		return err
	}

	queryParams := map[string]string{
		"chat_id":    tg.ChannelName,
		"photo":      game.Image,
		"parse_mode": "HTML",
		"caption":    buf.String(),
	}

	req := Request{
		Method: MethodGet,
		Name:   "sendPhoto",
		Params: &queryParams,
		Body:   nil,
	}

	Send(&req)
}

func (tg *Telegram) Post(game *epicgames.Game, silent bool) error {
	description := ""
	publisher := ""
	developer := ""
	price := ""
	availableRu := ""

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

	if !game.AvailableRu {
		availableRu = "\nНедоступно в РФ"
	}

	moscowLoc, err := time.LoadLocation("Europe/Moscow")

	if err != nil {
		return err
	}

	endDate := fmt.Sprintf(
		"\nИгра доступна бесплатно до %d %s, %s",
		game.Date.End.Day(),
		epicgames.GetMonth(game.Date.End.Month()),
		game.Date.End.In(moscowLoc).Format("15:04 MST"))

	messageText := JoinNotEmptyStrings(
		[]string{
			title,
			availableRu,
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

	getLogger().Println("Sending request to the Telegram API")
	resp, err := tg.Send(&req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	message := new(NewPostResponse)
	err = json.NewDecoder(resp.Body).Decode(&message)
	if err != nil {
		return err
	}

	return nil
}
