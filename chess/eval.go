package chess

const (
	Kingside = iota
	Queenside
)

type Position struct {
	King int                  // location of king
	Castle [2]bool            // castle availability
	Attack [128][]int         // attacking tiles
	Moves []Move              // actual, legal moves
	Score float32             // final score evaluation
}

func (pos *Position) Eval(b *Board, t Color) {
	pos.Moves = make([]Move, 0, 30)

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			tile := Tile(rank, file)

			if p := (*b)[tile]; p != nil && p.Color == t {
				if p.Kind == Pawn {
					p.
			}
		}
	}
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

	// add castle moves
	moves = append(moves, g.CastleMoves()...)

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
				Kind: kind,
			})

			if kind.Sliding() == false {
				break
			}
		}
	}

	return moves
}

func (g *Game) CastleMoves() []Move {
	tile := Tile(BackRank[g.Turn], 4)
	moves := make([]Move, 0, 2)

	// is the kingside castle available?
	if g.Castle & (Kingside << uint(g.Turn << 2)) != 0 {
		b := g.Board[tile + 1] == nil
		n := g.Board[tile + 2] == nil

		if b && n /* bishop and knight */ {
			moves = append(moves, Move{
				Origin: tile,
				Dest: tile + 2,
				Castle: Kingside,
				Kind: King,
			})
		}
	}

	// is the queenside castle available?
	if g.Castle & (Queenside << uint(g.Turn << 2)) != 0 {
		q := g.Board[tile - 1] == nil
		b := g.Board[tile - 2] == nil
		n := g.Board[tile - 3] == nil

		if q && b && n /* queen, bishop, knight */ {
			moves = append(moves, Move{
				Origin: tile,
				Dest: tile - 2,
				Castle: Queenside,
				Kind: King,
			})
		}
	}

	return moves
}
