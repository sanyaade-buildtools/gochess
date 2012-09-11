package chess

type Game struct {
	Position Board            // 0x88 board representation
	Turn Color                // whose turn it is
	King [2]int               // location of king pieces
	EnPassant int             // en passant availability
	Castles int               // castling availability
	HalfMove int              // pawn half moves
	Move int                  // current full move
}

type Move struct {
	Origin, Dest int          // where it is moving from and to
	Capture bool              // captured another piece
	Castle int                // castle move: Kingside or Queenside
	EnPassant bool            // was an en passant capture
	Pawn, Push, Promote bool  // pawn move, 2 space push, promotion
	Kind Kind                 // what was moved or promotion
}

const (
	Kingside = 1 + iota
	Queenside
)

func (g *Game) New() {
	g.Position.New()

	// initial state for a new game
	g.Turn = White
	g.King[White] = Tile(0, 4)
	g.King[Black] = Tile(7, 4)
	g.EnPassant = -1
	g.Castles = 15
	g.HalfMove = 0
	g.Move = 1
}

func (g *Game) IsLegalMove(move *Move) bool {
	if p := g.Position[move.Origin]; p == nil || p.Color != g.Turn {
		return false
	}

	// try making the move, which will update the position
	g.PerformMove(move)

	// cannot castle through (or out of) check
	switch move.Castle {
		case Kingside:
			for x := move.Origin; x < move.Dest; x++ {
				if g.IsAttacked(x) { return false }
			}
			break
		case Queenside:
			for x := move.Origin; x > move.Dest; x-- {
				if g.IsAttacked(x) { return false }
			}
			break
	}

	// cannot be in check after the move
	return g.IsAttacked(g.King[g.Turn]) == false
}

func (g *Game) IsAttacked(tile int) bool {
	piece := g.Position[tile]
	t := piece.Color
	opp := t.Opponent()

	// check pawn attacks to the left
	if l := tile - PieceDelta[Pawn][opp] - 1; Offboard(l) == false {
		if p := g.Position[l]; p != nil && p.Color == opp {
			return true
		}
	}

	// check pawn attack to the right
	if r := tile - PieceDelta[Pawn][opp] + 1; Offboard(r) == false {
		if p := g.Position[r]; p != nil && p.Color == opp {
			return true
		}
	}

	// check non-pawn piece types
	for kind := 1; kind < 6; kind++ {
		for _, d := range PieceDelta[kind] {
			capture := false

			// sliding pieces keep moving along that direction
			for x := tile + d; capture || Offboard(x) == false; x += d {
				if p := g.Position[x]; p != nil {
					if p.Color != t {
						return true
					}
					break
				}

				if Kind(kind).Sliding() == false {
					break
				}
			}
		}
	}

	return false
}

func (g *Game) PerformMove(move *Move) {
	// check for a castle move
	if move.Castle != 0 {
		rank := BackRank[g.Turn]

		switch move.Castle {
			case Kingside:
				g.Position.Move(Tile(rank, 4), Tile(rank, 6))
				g.Position.Move(Tile(rank, 7), Tile(rank, 5))
				break
			case Queenside:
				g.Position.Move(Tile(rank, 4), Tile(rank, 2))
				g.Position.Move(Tile(rank, 0), Tile(rank, 3))
				break
		}

		// performing castle means disabling castling
		g.DisableCastle(Kingside | Queenside)
	} else {
		// move the piece to the new position
		g.Position.Move(move.Origin, move.Dest)

		// pawn moves are special
		if move.Pawn {
			enPassant := move.Dest + PieceDelta[Pawn][g.Turn.Opponent()]

			switch {
				case move.EnPassant:
					g.Position.Remove(enPassant)
					break
				case move.Push:
					g.EnPassant = enPassant
					break
				case move.Promote:
					g.Position.Place(move.Dest, g.Turn, move.Kind)
					break
			}
		}

		// moving the rooks disables castling
		switch move.Origin {
			case Tile(BackRank[g.Turn], 0):
				g.DisableCastle(Queenside)
				break
			case Tile(BackRank[g.Turn], 7):
				g.DisableCastle(Kingside)
				break
		}
	}

	// disable en passant unless a pawn was pushed
	if move.Push == false {
		g.EnPassant = -1
	}

	// update the king's position if moved
	if move.Kind == King {
		g.King[g.Turn] = move.Dest

		// moving the king disabled all castling
		g.DisableCastle(Kingside | Queenside)
	}

	// record the move and switch whose turn it is
	if g.Turn = g.Turn.Opponent(); g.Turn == White {
		g.Move++
	}

	// update the half move counter
	if move.Pawn == true {
		g.HalfMove = 0
	} else {
		g.HalfMove++
	}
}

func (g *Game) DisableCastle(side int) {
	g.Castles &= ^(side << uint(g.Turn << 2))
}
