package epicgames

import (
	"errors"
	"time"
)

type gameType = byte
type giveawayTime = byte

const dateTimeFormat = time.RFC3339

const (
	gameTypeUnknown gameType = iota
	gameTypeCurrent
	gameTypeUpcoming
)

const (
	giveawayTimeUnknown giveawayTime = iota
	giveawayTimePast
	giveawayTimeNow
	giveawayTimeFuture
)

func getType(game *rawGame) (gameType, *promotionalOffer) {
	current := &game.Promotions.Current
	upcoming := &game.Promotions.Upcoming

	if len(*current) == 0 && len(*upcoming) == 0 {
		return gameTypeUnknown, nil
	}

	if len(*upcoming) == 0 && len(*current) > 0 {
		giveawayTime, promotionalOffer, err := getGiveawayTime(current)
		if err != nil {
			getLogger().Error(err.Error())
		} else {
			return selectGameType(giveawayTime), promotionalOffer
		}
	}

	if len(*current) == 0 && len(*upcoming) > 0 {
		giveawayTime, promotionalOffer, err := getGiveawayTime(upcoming)
		if err != nil {
			getLogger().Error(err.Error())
		} else {
			return selectGameType(giveawayTime), promotionalOffer
		}
	}

	if len(*current) > 0 && len(*upcoming) > 0 {
		giveawayTime, promotionalOffer, err := getGiveawayTime(current)
		if giveawayTime != giveawayTimeUnknown && err == nil {
			return selectGameType(giveawayTime), promotionalOffer
		}

		if err != nil {
			getLogger().Error(err.Error())
		}

		giveawayTime, promotionalOffer, err = getGiveawayTime(upcoming)
		if err == nil {
			return selectGameType(giveawayTime), promotionalOffer
		} else {
			getLogger().Error(err.Error())
		}
	}

	return gameTypeUnknown, nil
}

func getTime(offer promotionalOffer) (*time.Time, *time.Time, error) {
	if offer.StartDate == "" {
		return nil, nil, errors.New("empty start date")
	}

	if offer.EndDate == "" {
		return nil, nil, errors.New("empty end date")
	}

	startDate, err := time.Parse(dateTimeFormat, offer.StartDate)

	if err != nil {
		return &startDate, nil, err
	}

	endDate, err := time.Parse(dateTimeFormat, offer.EndDate)

	if err != nil {
		return nil, &endDate, err
	}

	return &startDate, &endDate, nil
}

func selectGameType(giveawayTime giveawayTime) gameType {
	switch giveawayTime {
	case giveawayTimeNow:
		return gameTypeCurrent
	case giveawayTimeFuture:
		return gameTypeUpcoming
	default:
		return gameTypeUnknown
	}
}

func getGiveawayTime(promotions *promotions) (giveawayTime, *promotionalOffer, error) {
	now := time.Now()

	gaTime := giveawayTimeUnknown

	promotionalOffers := (*promotions)[0].PromotionalOffers

	var neededPromotionalOffer promotionalOffer

	var startDate, endDate *time.Time
	var firstValidIndex int
	var err error

	for index, promotionalOffer := range promotionalOffers {
		startDate, endDate, err = getTime(promotionalOffer)
		if err != nil {
			getLogger().Error(err.Error())
			continue
		}

		if !isGiveawayItem(&promotionalOffer) {
			continue
		}

		firstValidIndex = index
		neededPromotionalOffer = promotionalOffer
		break
	}

	if len(promotionalOffers) > firstValidIndex {
		for _, promotionalOffer := range promotionalOffers[firstValidIndex+1:] {
			tmpStartDate, tmpEndDate, err := getTime(promotionalOffer)
			if err != nil {
				getLogger().Error(err.Error())
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

	if startDate == nil {
		return giveawayTimeUnknown, nil, errors.New("empty start date")
	}

	if endDate == nil {
		return giveawayTimeUnknown, nil, errors.New("empty end date")
	}

	isRelevant := isRelevantDate(startDate)
	isForGiveaway := isGiveawayItem(&neededPromotionalOffer)

	if !isRelevant || !isForGiveaway {
		return giveawayTimeUnknown, nil, nil
	}

	startBeforeNow := startDate.Before(now)
	startAfterNow := startDate.After(now)
	endAfterNow := endDate.After(now)

	switch {
	case startBeforeNow && endDate.Before(now):
		gaTime = giveawayTimePast
	case startBeforeNow && endAfterNow:
		gaTime = giveawayTimeNow
	case startAfterNow && endAfterNow:
		gaTime = giveawayTimeFuture
	}

	return gaTime, &neededPromotionalOffer, nil
}

func isGiveawayItem(offer *promotionalOffer) bool {
	return offer.DiscountSetting.DiscountType == "PERCENTAGE" && offer.DiscountSetting.DiscountPercentage == 0
}

func isRelevantDate(date *time.Time) bool {
	return date == nil ||
		!(date.Before(time.Now().AddDate(0, 0, -8)) ||
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
