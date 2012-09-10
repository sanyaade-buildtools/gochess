package chess

func Eval(pos *Position, t Color) []*Move {
	var moves []*Move

	c := make(chan *Move)

	go PseudoLegalMoves(c, pos, t)

	for move := range c {
		moves = append(moves, move)
	}

	return moves
}

func PseudoLegalMoves(c chan *Move, pos *Position, t Color) {
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			tile := Tile(rank, file)

			// find moves for pieces matching the color
			if p := pos.Board[tile]; p != nil && p.Color == t {
				if p.Kind == Pawn {
					PawnMoves(c, pos, tile)
				} else {
					NonPawnMoves(c, pos, tile, p.Kind)
				}
			}
		}
	}

	// get available castle moves
	CastleMoves(c, pos, t)

	// done collecting moves
	close(c)
}

func PawnMoves(c chan *Move, pos *Position, tile int) {
	t := pos.Board[tile].Color
	d := PieceDelta[Pawn][t]
	x := tile + d

	// advance forward once (can't be off board)
	if pos.Board[x] == nil {
		c <- &Move{
			Origin: tile,
			Dest: x,
			Pawn: true,
			Promote: Rank(x) == BackRank[t.Opponent()],
		}

		// try pushing the pawn?
		if Rank(tile) == PawnRank[t] {
			if pos.Board[x + d] == nil {
				c <- &Move{
					Origin: tile,
					Dest: x + d,
					Pawn: true,
					Push: true,
				}
			}
		}
	}

	// capturing and en passant
	for _, i := range [2]int{ -1, 1 } {
		if Offboard(x + i) {
			continue
		}

		// en passant capture?
		if tile + d + i == pos.EnPassant {
			c <- &Move{
				Origin: tile,
				Dest: x + i,
				Pawn: true,
				Capture: true,
				EnPassant: true,
			}
		} else {
			p := pos.Board[x + i]

			if p != nil && p.Color != t {
				c <- &Move{
					Origin: tile,
					Dest: x + i,
					Pawn: true,
					Capture: true,
					Promote: Rank(x + i) == BackRank[t.Opponent()],
				}
			}
		}
	}
}

func NonPawnMoves(c chan *Move, pos *Position, tile int, kind Kind) {
	t := pos.Board[tile].Color

	// move in all allowed directions of travel
	for _, d := range PieceDelta[kind] {
		capture := false

		// sliding pieces keep moving along that direction
		for x := tile + d; capture || Offboard(x) == false; x += d {
			if p := pos.Board[x]; p != nil {
				if p.Color != t {
					capture = true
				} else {
					break
				}
			}

			c <- &Move{
				Origin: tile,
				Dest: x,
				Capture: capture,
				Kind: kind,
			}

			if kind.Sliding() == false {
				break
			}
		}
	}
}

func CastleMoves(c chan *Move, pos *Position, t Color) {
	tile := Tile(BackRank[t], 4)

	// is the kingside castle available?
	if pos.Castles & (Kingside << uint(t << 2)) != 0 {
		b := pos.Board[tile + 1] == nil
		n := pos.Board[tile + 2] == nil

		if b && n /* bishop and knight */ {
			c <- &Move{
				Origin: tile,
				Dest: tile + 2,
				Castle: Kingside,
				Kind: King,
			}
		}
	}

	// is the queenside castle available?
	if pos.Castles & (Queenside << uint(t << 2)) != 0 {
		q := pos.Board[tile - 1] == nil
		b := pos.Board[tile - 2] == nil
		n := pos.Board[tile - 3] == nil

		if q && b && n /* queen, bishop, knight */ {
			c <- &Move{
				Origin: tile,
				Dest: tile - 2,
				Castle: Queenside,
				Kind: King,
			}
		}
	}
}
