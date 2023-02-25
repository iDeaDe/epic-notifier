package epicgames

import (
	"time"
)

type GameType = byte
type GiveawayTime = byte

const DateTimeFormat = time.RFC3339

const (
	UnknownType GameType = iota
	Current
	Upcoming
)

const (
	UnknownTime GiveawayTime = iota
	Past
	Now
	Future
)

func GetType(game *RawGame) GameType {
	current := &game.Promotions.Current
	upcoming := &game.Promotions.Upcoming

	if len(*current) == 0 && len(*upcoming) == 0 {
		return UnknownType
	}

	if len(*current) == 0 && len(*upcoming) > 0 {
		giveawayTime, err := getGiveawayTime(upcoming)
		if err == nil {
			return selectGameType(giveawayTime)
		}
	}

	if len(*upcoming) == 0 && len(*current) > 0 {
		giveawayTime, err := getGiveawayTime(current)
		if err == nil {
			return selectGameType(giveawayTime)
		}
	}

	if len(*current) > 0 && len(*upcoming) > 0 {
		giveawayTime, err := getGiveawayTime(upcoming)
		if err == nil {
			return selectGameType(giveawayTime)
		}

		giveawayTime, err = getGiveawayTime(current)
		if err == nil {
			return selectGameType(giveawayTime)
		}
	}

	return UnknownType
}

func GetTime(offer PromotionalOffer) (*time.Time, *time.Time, error) {
	startDate, err := time.Parse(DateTimeFormat, offer.StartDate)

	if err != nil {
		return &startDate, nil, err
	}

	endDate, err := time.Parse(DateTimeFormat, offer.EndDate)

	if err != nil {
		return nil, &endDate, err
	}

	return &startDate, &endDate, nil
}

func selectGameType(giveawayTime GiveawayTime) GameType {
	switch giveawayTime {
	case Now:
		return Current
	case Future:
		return Upcoming
	}

	return UnknownType
}

func getGiveawayTime(promotions *Promotions) (GiveawayTime, error) {
	now := time.Now()
	startDate, endDate, err := GetTime((*promotions)[0].PromotionalOffers[0])

	if err != nil {
		return UnknownTime, err
	}

	gaTime := UnknownTime

	startBeforeNow := startDate.Before(now)
	startAfterNow := startDate.After(now)
	endAfterNow := endDate.After(now)

	if startDate.Before(time.Now().AddDate(0, 0, -8)) ||
		startDate.After(now.AddDate(0, 0, 8)) {
		return UnknownType, nil
	}

	switch {
	case startBeforeNow && endDate.Before(now):
		gaTime = Past
		break
	case startBeforeNow && endAfterNow:
		gaTime = Now
		break
	case startAfterNow && endAfterNow:
		gaTime = Future
	}

	return gaTime, nil
}
