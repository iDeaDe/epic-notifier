package epicgames

type GameType = byte

const (
	Unknown GameType = iota
	Current
	Upcoming
)

func GetType(game *RawGame) GameType {
	current := &game.Promotions.Current
	upcoming := &game.Promotions.Upcoming

	if *current == nil && *upcoming == nil || len(*current) == 0 && len(*upcoming) == 0 {
		return Unknown
	}

	if (*current == nil || len(*current) == 0) && *upcoming != nil && len(*upcoming) > 0 {
		return Upcoming
	}

	if (*upcoming == nil || len(*upcoming) == 0) && *current != nil && len(*current) > 0 {
		return Current
	}

	if len(*current) > 0 && len(*upcoming) > 0 {
		return Current
	}

	return Unknown
}
