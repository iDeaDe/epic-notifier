package epicgames

import (
	"time"
)

type GameType = byte

const (
	Unknown GameType = iota
	Current
	Upcoming
)

func GetType(game *RawGame) GameType {
	current := &game.Promotions.Current
	upcoming := &game.Promotions.Upcoming

	if len(*current) == 0 && len(*upcoming) == 0 {
		return Unknown
	}

	if len(*current) == 0 && len(*upcoming) > 0 {
		available, err := isAvailable(upcoming)
		if err == nil && available {
			return Current
		}

		return Upcoming
	}

	if len(*upcoming) == 0 && len(*current) > 0 {
		available, err := isAvailable(current)
		if err == nil && !available {
			return Upcoming
		}

		return Current
	}

	if len(*current) > 0 && len(*upcoming) > 0 {
		available, err := isAvailable(upcoming)
		if err == nil && !available {
			return Upcoming
		}

		available, err = isAvailable(current)
		if err == nil && available {
			return Current
		}
	}

	return Unknown
}

func GetTime(offer PromotionalOffer) (*time.Time, *time.Time, error) {
	startDate, err := time.Parse(time.RFC3339, offer.StartDate)

	if err != nil {
		return &startDate, nil, err
	}

	endDate, err := time.Parse(time.RFC3339, offer.EndDate)

	if err != nil {
		return nil, &endDate, err
	}

	return &startDate, &endDate, nil
}

func isAvailable(promotions *Promotions) (bool, error) {
	for _, promotion := range *promotions {
		startDate, endDate, err := GetTime(promotion["promotionalOffers"][0])

		if err != nil {
			return false, err
		}

		return startDate.Before(time.Now()) && endDate.After(time.Now()), nil
	}

	return false, nil
}
