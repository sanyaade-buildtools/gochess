package chess

func (g *Game) LegalMoves(pseudoMoves chan *Move) chan *Move {
	c := make(chan *Move, 60)

	go func() {
		for move := range pseudoMoves {
			c <- move // TODO: validate the move is 100% legal
		}

		// now have full set of moves
		close(c)
	}()

	return c
}

func (g *Game) PseudoLegalMoves() chan *Move {
	c := make(chan *Move, 60)

	go func() {
		for rank := 0; rank < 8; rank++ {
			for file := 0; file < 8; file++ {
				tile := Tile(rank, file)

				// find moves for pieces matching the color
				if p := g.Position[tile]; p != nil && p.Color == g.Turn {
					if p.Kind == Pawn {
						pawnMoves(c, g, tile)
					} else {
						nonPawnMoves(c, g, tile, p.Kind)
					}
				}
			}
		}

		// get available castle moves
		castleMoves(c, g)

		// done collecting moves
		close(c)
	}()

	return c
}

func pawnMoves(c chan *Move, g *Game, tile int) {
	d := PieceDelta[Pawn][g.Turn]
	back := BackRank[g.Turn.Opponent()]
	x := tile + d

	// advance forward once (can't be off board)
	if g.Position[x] == nil {
		c <- &Move{
			Origin: tile,
			Dest: x,
			Pawn: true,
			Promote: Rank(x) == back,
		}

		// try pushing the pawn?
		if Rank(tile) == PawnRank[g.Turn] {
			if g.Position[x + d] == nil {
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
		if tile + d + i == g.EnPassant {
			c <- &Move{
				Origin: tile,
				Dest: x + i,
				Pawn: true,
				Capture: true,
				EnPassant: true,
			}
		} else {
			p := g.Position[x + i]

			if p != nil && p.Color != g.Turn {
				c <- &Move{
					Origin: tile,
					Dest: x + i,
					Pawn: true,
					Capture: true,
					Promote: Rank(x + i) == back,
				}
			}
		}
	}
}

func nonPawnMoves(c chan *Move, g *Game, tile int, kind Kind) {
	for _, d := range PieceDelta[kind] {
		capture := false

		// sliding pieces keep moving along that direction
		for x := tile + d; capture || Offboard(x) == false; x += d {
			if p := g.Position[x]; p != nil {
				if p.Color != g.Turn {
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

func castleMoves(ch chan *Move, g *Game) {
	tile := Tile(BackRank[g.Turn], 4)

	// is the kingside castle available?
	if g.Castles & (Kingside << uint(g.Turn << 2)) != 0 {
		b := g.Position[tile + 1] == nil
		n := g.Position[tile + 2] == nil

		if b && n /* bishop and knight */ {
			ch <- &Move{
				Origin: tile,
				Dest: tile + 2,
				Castle: Kingside,
				Kind: King,
			}
		}
	}

	// is the queenside castle available?
	if g.Castles & (Queenside << uint(g.Turn << 2)) != 0 {
		q := g.Position[tile - 1] == nil
		b := g.Position[tile - 2] == nil
		n := g.Position[tile - 3] == nil

		if q && b && n /* queen, bishop, knight */ {
			ch <- &Move{
				Origin: tile,
				Dest: tile - 2,
				Castle: Queenside,
				Kind: King,
			}
		}
	}
}
