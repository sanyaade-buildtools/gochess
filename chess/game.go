package chess

type Game struct {
	Position Position         // the current position
	History []Move            // history of all moves made
}

func (g *Game) NewGame() {
	g.Position.Setup()
	g.History = make([]Move, 0, 50)
}
