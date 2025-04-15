package game

type game struct {
	players map[string]Player
}

func NewGame() *game {
	return &game{
		players: make(map[string]Player, 0),
	}
}

