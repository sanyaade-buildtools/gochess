package chess

type Game struct {
	Board Board               // position representation
	Turn Color                // white or black
	Castle int                // what castle moves are available
	EnPassant int             // tile en passant can be performed into
	HalfMove, Move int        // pawn half moves and full move count
}

type Move struct {
	Origin, Dest int          // where it is moving from and to
	Capture bool              // captured another piece
	Castle int                // castle move (0=not a castle)
	EnPassant bool            // was an en passant capture
	Pawn, Push, Promote bool  // pawn move, 2 space push, promotion
	Kind Kind                 // promote piece, ...
}

const (
	Kingside = 1 + iota
	Queenside
)

var PieceDelta = [6][]int{
	[]int{ 16, -16 },
	[]int{ -15, 15, -17, 17 },
	[]int{ -31, -33, -14, 18, -18, 14, 31, 33 },
	[]int{ -1, 1, -16, 16, -17, 15, -15, 17 },
	[]int{ -1, 1, -16, 16 },
	[]int{ -1, 1, -16, 16, -17, 15, -15, 17 },
}

func (g *Game) NewGame() {
	g.Board.Setup()
	g.Turn = White
	g.Castle = 15
	g.EnPassant = -1
	g.HalfMove = 0
	g.Move = 0
}

func (g *Game) PseudoLegalMoves() []Move {
	var moves []Move

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			tile := Tile(rank, file)

			// must have a piece belonging to the player
			if p := g.Board[tile]; p != nil && p.Color == g.Turn {
				if p.Kind == Pawn {
					moves = append(moves, g.PawnMoves(tile)...)
				} else {
					moves = append(moves, g.NonPawnMoves(tile, p.Kind)...)
				}
			}
		}
	}

	// TODO: add castle moves

	return moves
}

func (g *Game) PawnMoves(tile int) []Move {
	moves := make([]Move, 0, 4)

	// the singular direction of travel for a pawn
	d := PieceDelta[Pawn][g.Turn]
	pos := tile + d

	// advance forward once (can't be off board)
	if g.Board[pos] == nil {
		moves = append(moves, Move{
			Origin: tile,
			Dest: pos,
			Pawn: true,
			Promote: Rank(pos) == BackRank[g.Turn.Opponent()],
		})

		// try pushing the pawn?
		if Rank(tile) == PawnRank[g.Turn] {
			if g.Board[pos + d] == nil {
				moves = append(moves, Move{
					Origin: tile,
					Dest: pos + d,
					Pawn: true,
					Push: true,
				})
			}
		}
	}

	// capturing and en passant
	for _, i := range [2]int{ -1, 1 } {
		if Offboard(pos + i) {
			continue
		}

		// en passant capture?
		if tile + d + i == g.EnPassant {
			moves = append(moves, Move{
				Origin: tile,
				Dest: pos + i,
				Pawn: true,
				Capture: true,
				EnPassant: true,
			})
		} else {
			p := g.Board[pos + i]

			if p != nil && p.Color != g.Turn {
				moves = append(moves, Move{
					Origin: tile,
					Dest: pos + i,
					Pawn: true,
					Capture: true,
					Promote: Rank(pos + i) == BackRank[g.Turn.Opponent()],
				})
			}
		}
	}

	return moves
}

func (g *Game) NonPawnMoves(tile int, kind Kind) []Move {
	moves := make([]Move, 0, 28)

	// traverse in every direction the piece can move
	for _, d := range PieceDelta[kind] {
		capture := false

		// sliding pieces keep moving along that direction
		for pos := tile + d; capture || Offboard(pos) == false ; pos += d {
			if p := g.Board[pos]; p != nil {
				if p.Color != g.Turn {
					capture = true
				} else {
					break
				}
			}

			moves = append(moves, Move{
				Origin: tile,
				Dest: pos,
				Capture: capture,
			})

			if kind.Sliding() == false {
				break
			}
		}
	}

	return moves
}
