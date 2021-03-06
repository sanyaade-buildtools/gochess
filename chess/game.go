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
	Check int                 // check and/or mate
	EnPassant bool            // was an en passant capture
	Pawn, Push, Promote bool  // pawn move, 2 space push, promotion
	Kind Kind                 // what was moved or promotion
}

// Castle availability is a 4-bit nibble with 2 bits for white
// and 2 bits for black. To test the availability of a particular
// castle move for a player, test (Side << (Color << 2)).

const (
	Kingside = 1 + iota
	Queenside
)

const (
	Check = 1 + iota
	Mate
)

func NewGame() *Game {
	g := new(Game)

	// setup the new board
	g.Position.New()

	// initial state for a new game
	g.Turn = White
	g.King[White] = Tile(0, 4)
	g.King[Black] = Tile(7, 4)
	g.EnPassant = -1
	g.Castles = 15
	g.HalfMove = 0
	g.Move = 1

	return g
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
