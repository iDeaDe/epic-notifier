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

func GetType(game *RawGame) (GameType, *PromotionalOffer) {
	current := &game.Promotions.Current
	upcoming := &game.Promotions.Upcoming

	if len(*current) == 0 && len(*upcoming) == 0 {
		return UnknownType, nil
	}

	if len(*upcoming) == 0 && len(*current) > 0 {
		giveawayTime, promotionalOffer, err := getGiveawayTime(current)
		if err != nil {
			getLogger().Println(err)
		} else {
			return selectGameType(giveawayTime), promotionalOffer
		}
	}

	if len(*current) == 0 && len(*upcoming) > 0 {
		giveawayTime, promotionalOffer, err := getGiveawayTime(upcoming)
		if err != nil {
			getLogger().Println(err)
		} else {
			return selectGameType(giveawayTime), promotionalOffer
		}
	}

	if len(*current) > 0 && len(*upcoming) > 0 {
		giveawayTime, promotionalOffer, err := getGiveawayTime(current)
		if giveawayTime != UnknownTime && err == nil {
			return selectGameType(giveawayTime), promotionalOffer
		}

		if err != nil {
			getLogger().Println(err)
		}

		giveawayTime, promotionalOffer, err = getGiveawayTime(upcoming)
		if err == nil {
			return selectGameType(giveawayTime), promotionalOffer
		} else {
			getLogger().Println(err)
		}
	}

	return UnknownType, nil
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

func getGiveawayTime(promotions *Promotions) (GiveawayTime, *PromotionalOffer, error) {
	now := time.Now()

	gaTime := UnknownTime

	promotionalOffers := (*promotions)[0].PromotionalOffers

	var neededPromotionalOffer PromotionalOffer

	var startDate, endDate *time.Time
	var firstValidIndex int
	var err error

	for index, promotionalOffer := range promotionalOffers {
		startDate, endDate, err = GetTime(promotionalOffer)
		if err != nil {
			getLogger().Println(err)
			continue
		}

		firstValidIndex = index
		neededPromotionalOffer = promotionalOffer
		break
	}

	if len(promotionalOffers) > firstValidIndex {
		for _, promotionalOffer := range promotionalOffers[firstValidIndex+1:] {
			tmpStartDate, tmpEndDate, err := GetTime(promotionalOffer)
			if err != nil {
				getLogger().Println(err)
				continue
			}

			if !isRelevantDate(tmpStartDate) || !isGiveawayItem(&promotionalOffer) {
				continue
			}

			if tmpStartDate.Before(*startDate) || tmpStartDate.Equal(*startDate) {
				startDate = tmpStartDate
				endDate = tmpEndDate
				neededPromotionalOffer = promotionalOffer
			}
		}
	}

	if !isRelevantDate(startDate) || !isGiveawayItem(&neededPromotionalOffer) {
		return UnknownTime, nil, nil
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

	return gaTime, &neededPromotionalOffer, nil
}

func isGiveawayItem(offer *PromotionalOffer) bool {
	return offer.DiscountSetting.DiscountType == "PERCENTAGE" && offer.DiscountSetting.DiscountPercentage == 0
}

func isRelevantDate(date *time.Time) bool {
	return !(date.Before(time.Now().AddDate(0, 0, -8)) ||
		date.After(time.Now().AddDate(0, 0, 8)))
}

func filterNextGames(ga *Giveaway) {
	var tmpGames []Game

	for _, game := range ga.NextGames {
		if game.Date.Start.Before(ga.Next) || game.Date.Start.Equal(ga.Next) {
			tmpGames = append(tmpGames, game)
		}
	}

	ga.NextGames = tmpGames
}
