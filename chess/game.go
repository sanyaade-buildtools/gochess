package chess

type Game struct {
	Position Position         // the current position
	History []Move            // history of all moves made
}

type Move struct {
	Origin, Dest int          // where it is moving from and to
	Capture bool              // captured another piece
	Castle int                // castle move: Kingside or Queenside
	EnPassant bool            // was an en passant capture
	Pawn, Push, Promote bool  // pawn move, 2 space push, promotion
	Kind Kind                 // what was moved or promotion
}

func (g *Game) Init() {
	g.History = make([]Move, 0, 50)
	g.Position.NewGame()
}
