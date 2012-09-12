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

func (g *Game) LegalMoves() []*Move {
	moves := make([]*Move, 0, 30)
	c := g.PseudoLegalMoves()

	// collect all pseudo-legal moves that are 100% legal
	for move := range c {
		if g.IsLegalMove(move) {
			moves = append(moves, move)
		}
	}

	return moves
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
		d := 1

		// make sure the castle move is available
		if g.Castles & (move.Castle << uint(g.Turn << 2)) == 0 {
			return false
		}

		// queenside castles move down in file
		if move.Castle != Kingside {
			d = -d
		}

		// cannot move through check
		for i := 0; i < 3; i++ {
			
			if g.InCheck(g.King[g.Turn] + i * d) {
				return false
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
