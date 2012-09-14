package chess

// these are all the methods that each side can use to attack with
var AttackTable = [2]map[int][][]int{
	map[int][][]int{
		(1 << uint(Bishop)) | (1 << uint(Queen)): [][]int{
			[]int{ 15, 30, 45, 60, 75, 90, 105 }, // up left
			[]int{ 17, 34, 51, 68, 85, 102, 119 }, // up right
			[]int{ -15, -30, -45, -60, -75, -90, -105 }, // down left
			[]int{ -17, -34, -51, -68, -85, -102, -119 }, // down right
		},
		(1 << uint(Rook)) | (1 << uint(Queen)): [][]int{
			[]int{ 16, 32, 48, 64, 80, 96, 112 }, // up
			[]int{ 1, 2, 3, 4, 5, 6, 7 }, // right
			[]int{ -16, -32, -48, -64, -80, -96, -112 }, // down
			[]int{ -1, -2, -3, -4, -5, -6, -7 }, // left
		},
		(1 << uint(Knight)): [][]int{ PieceDelta[Knight] },
		(1 << uint(King)): [][]int{ PieceDelta[King] },
		(1 << uint(Pawn)): [][]int{
			[]int{ 
				PieceDelta[Pawn][White] - 1,
				PieceDelta[Pawn][White] + 1,
			},
		},
	},
	map[int][][]int{
		(1 << uint(Bishop)) | (1 << uint(Queen)): [][]int{
			[]int{ 15, 30, 45, 60, 75, 90, 105 }, // up left
			[]int{ 17, 34, 51, 68, 85, 102, 119 }, // up right
			[]int{ -15, -30, -45, -60, -75, -90, -105 }, // down left
			[]int{ -17, -34, -51, -68, -85, -102, -119 }, // down right
		},
		(1 << uint(Rook)) | (1 << uint(Queen)): [][]int{
			[]int{ 16, 32, 48, 64, 80, 96, 112 }, // up
			[]int{ 1, 2, 3, 4, 5, 6, 7 }, // right
			[]int{ -16, -32, -48, -64, -80, -96, -112 }, // down
			[]int{ -1, -2, -3, -4, -5, -6, -7 }, // left
		},
		(1 << uint(Knight)): [][]int{ PieceDelta[Knight] },
		(1 << uint(King)): [][]int{ PieceDelta[King] },
		(1 << uint(Pawn)): [][]int{
			[]int{ 
				PieceDelta[Pawn][Black] - 1,
				PieceDelta[Pawn][Black] + 1,
			},
		},
	},
}

func (g *Game) IsLegalMove(move *Move) bool {
	if Offboard(move.Origin) || Offboard(move.Dest) {
		return false
	}

	p := g.Position[move.Origin]
	x := g.Position[move.Dest]

	// the piece exists and is owned by the current player
	if p == nil || p.Color != g.Turn {
		return false
	}

	if move.Castle != 0 {
		if g.Castles & (move.Castle << uint(g.Turn << 2)) == 0 {
			return false
		}

		// queenside castles move down in file
		if move.Castle == Kingside {
			for i := 0; i < 3; i++ {
				if g.InCheck(g.King[g.Turn] + i) {
					return false
				}
			}
		} else {
			for i := 0; i < 3; i++ {
				if g.InCheck(g.King[g.Turn] - i) {
					return false
				}
			}
		}

		return true
	}

	if move.Pawn {
		if p.Kind != Pawn {
			return false
		}

		// can only push pawns on the pawn rank
		if move.Push && Rank(move.Origin) != PawnRank[g.Turn] {
			return false
		}

		// make sure en passant is valid
		if move.EnPassant && move.Dest != g.EnPassant {
			return false
		}
	}

	// undo
	defer func() {
		g.Position[move.Origin] = p
		g.Position[move.Dest] = x
	}()

	// make move
	g.Position[move.Origin] = nil
	g.Position[move.Dest] = p

	// verify (if the king moved) it isn't in check
	if move.Origin == g.King[g.Turn] {
		return g.InCheck(move.Dest) == false
	}

	// verify the king isn't in check
	return g.InCheck(g.King[g.Turn]) == false
}

func (g *Game) InCheck(tile int) bool {
	for pieces, attacks := range AttackTable[g.Turn.Opponent()] {
		for _, direction := range attacks {
			for _, delta := range direction {
				if Offboard(tile + delta) {
					break
				}

				// check to see if a piece is there
				if p := g.Position[tile]; p != nil {
					bit := 1 << uint(p.Kind)

					// my piece or wrong piece
					if p.Color == g.Turn  || (pieces & bit == 0) {
						break
					}

					// enenmy piece of the right type
					return true
				}
			}
		}
	}

	return false
}

func (g *Game) CollectMoves() []*Move {
	pseudoMoves := make(chan *Move)
	moves := make([]*Move, 0, 30)

	go func() {
		for rank := 0; rank < 8; rank++ {
			for file := 0; file < 8; file++ {
				tile := Tile(rank, file)

				// only collect moves for this player
				if p := g.Position[tile]; p != nil && p.Color == g.Turn {
					if p.Kind == Pawn {
						g.PawnMoves(pseudoMoves, tile)
					} else {
						g.NonPawnMoves(pseudoMoves, tile, p.Kind)
					}
				}
			}
		}

		// add castling moves
		g.CastleMoves(pseudoMoves)

		// all pseudo legal moves have been collected
		close(pseudoMoves)
	}()

	// filter legal moves from the pseudo legal ones
	for move := range pseudoMoves {
		if g.IsLegalMove(move) {
			moves = append(moves, move)
		}
	}

	return moves
}

func (g *Game) PawnMoves(ch chan *Move, tile int) {
	d := PieceDelta[Pawn][g.Turn]
	x := tile + d

	// advance forward once (can't be off board)
	if g.Position[x] == nil {
		ch <- &Move{
			Origin: tile,
			Dest: x,
			Pawn: true,
			Promote: Rank(x) == BackRank[g.Turn.Opponent()],
		}

		// try pushing the pawn?
		if Rank(tile) == PawnRank[g.Turn] {
			if g.Position[x + d] == nil {
				ch <- &Move{
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
			ch <- &Move{
				Origin: tile,
				Dest: x + i,
				Pawn: true,
				Capture: true,
				EnPassant: true,
			}
		} else {
			p := g.Position[x + i]

			if p != nil && p.Color != g.Turn {
				ch <- &Move{
					Origin: tile,
					Dest: x + i,
					Pawn: true,
					Capture: true,
					Promote: Rank(x + i) == BackRank[g.Turn.Opponent()],
				}
			}
		}
	}
}

func (g *Game) NonPawnMoves(ch chan *Move, tile int, kind Kind) {
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

			ch <- &Move{
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

func (g *Game) CastleMoves(ch chan *Move) {
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
