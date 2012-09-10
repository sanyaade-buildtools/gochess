package chess

const (
	Kingside = 1 + iota
	Queenside
)

// Position.Castle is a 4-bit nibble, where each 2 bits represents
// the kingside and queenside availability. To get the bit value for
// a given player = CastleSide << (PlayerColor << 2)

type Position struct {
	Board Board               // 0x88 representation
	Turn Color                // whose turn it is
	King [2]int               // location of the kings
	EnPassant int             // tile en passant can be performed into
	Castles int               // what castle moves are available
	HalfMove int              // pawn half moves
	Start int                 // what move this game started from
}

type Move struct {
	Origin, Dest int          // where it is moving from and to
	Capture bool              // captured another piece
	Castle int                // castle move: Kingside or Queenside
	EnPassant bool            // was an en passant capture
	Pawn, Push, Promote bool  // pawn move, 2 space push, promotion
	Kind Kind                 // what was moved or promotion
}

func (m *Move) IsAttack() bool {
	return m.Castle == 0 && (m.Pawn == false || m.Capture == true)
}

func (pos *Position) Setup() {
	pos.Turn = White
	pos.Board.Setup()
	pos.King[White] = Tile(0, 4)
	pos.King[Black] = Tile(7, 4)
	pos.EnPassant = -1
	pos.Castles = 15
	pos.HalfMove = 0
	pos.Start = 1
}

func (pos *Position) PerformMove(move Move) {
	if move.Castle != 0 {
		pos.PerformCastle(move.Castle)
	} else {
		pos.Board.Move(move.Origin, move.Dest)

		// pawn moves are special
		if move.Pawn {
			enPassant := move.Dest + PieceDelta[Pawn][pos.Turn.Opponent()]

			switch {
				case move.EnPassant:
					pos.Board.Remove(enPassant)
					break
				case move.Push:
					pos.EnPassant = enPassant
					break
				case move.Promote:
					pos.Board.Place(move.Dest, pos.Turn, move.Kind)
					break
			}
		}

		// moving the rooks disables castling
		switch move.Origin {
			case Tile(BackRank[pos.Turn], 0):
				pos.DisableCastle(Queenside)
				break
			case Tile(BackRank[pos.Turn], 7):
				pos.DisableCastle(Kingside)
				break
		}
	}

	// update the king's position if moved
	if move.Kind == King {
		pos.King[pos.Turn] = move.Dest
		pos.DisableCastle(Kingside | Queenside)
	}

	// record the move and switch whose turn it is
	pos.Turn = pos.Turn.Opponent()

	// update the half move counter
	if move.Pawn == true {
		pos.HalfMove = 0
	} else {
		pos.HalfMove++
	}
}

func (pos *Position) PerformCastle(castle int) {
	rank := BackRank[pos.Turn]

	switch castle {
		case Kingside:
			pos.Board.Move(Tile(rank, 4), Tile(rank, 6))
			pos.Board.Move(Tile(rank, 7), Tile(rank, 5))
			break
		case Queenside:
			pos.Board.Move(Tile(rank, 4), Tile(rank, 2))
			pos.Board.Move(Tile(rank, 0), Tile(rank, 3))
			break
	}

	pos.DisableCastle(Kingside | Queenside)
}

func (pos *Position) DisableCastle(side int) {
	pos.Castles &= ^(side << uint(pos.Turn << 2))
}
