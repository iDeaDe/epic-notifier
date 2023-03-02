package epicgames

import (
	"sort"
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

	gaTime := UnknownTime

	promotionalOffers := &((*promotions)[0].PromotionalOffers)

	sort.Slice(*promotionalOffers, func(i, j int) bool {
		startTimeI, _, err1 := GetTime((*promotionalOffers)[i])
		startTimeJ, _, err2 := GetTime((*promotionalOffers)[j])

		if err1 != nil || err2 != nil {
			return err1 != nil && err2 == nil
		}

		return startTimeI.Before(*startTimeJ)
	})

	for _, promotionalOffer := range (*promotions)[0].PromotionalOffers {
		startDate, endDate, err := GetTime(promotionalOffer)
		if err != nil {
			continue
		}

		if startDate.Before(time.Now().AddDate(0, 0, -8)) ||
			startDate.After(now.AddDate(0, 0, 8)) {
			continue
		}

		startBeforeNow := startDate.Before(now)
		startAfterNow := startDate.After(now)
		endAfterNow := endDate.After(now)

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
	}

	return gaTime, nil
}

func filterGames(ga *Giveaway) {
	for index, game := range ga.NextGames {
		if game.Date.Start.After(ga.Next) {
			ga.NextGames = removeGamesByIndex(ga.NextGames, index)
		}
	}
}
