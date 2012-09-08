package chess

type Position struct {
	Board Board               // 0x88 representation
	Turn Color                // whose turn it is
	King [2]int               // position of the kings
	EnPassant int             // tile en passant can be performed into
	Castles int               // what castle moves are available
	HalfMove int              // pawn half moves
	Start int                 // what move this game started from
}

// Position.Castle is a 4-bit nibble, where each 2 bits represents
// the kingside and queenside availability. To get the bit value for
// a given player = Side << (Color << 2)

const (
	Kingside = 1 + iota
	Queenside
)

func (pos *Position) Init() {
	pos.Turn = White
	pos.EnPassant = -1
	pos.Castles = 15
	pos.HalfMove = 0
	pos.Start = 1
}

func (pos *Position) NewGame() {
	pos.Init()
	pos.Board.Setup()
	pos.King[White] = Tile(0, 4)
	pos.King[Black] = Tile(7, 4)
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

func (pos *Position) PseudoLegalMoves() []Move {
	var moves []Move

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			tile := Tile(rank, file)

			// must have a piece belonging to the player
			if p := pos.Board[tile]; p != nil && p.Color == pos.Turn {
				if p.Kind == Pawn {
					moves = append(moves, pos.PawnMoves(tile)...)
				} else {
					moves = append(moves, pos.NonPawnMoves(tile, p.Kind)...)
				}
			}
		}
	}

	// add castle moves when allowed
	moves = append(moves, pos.CastleMoves()...)

	return moves
}

func (pos *Position) PawnMoves(tile int) []Move {
	moves := make([]Move, 0, 4)

	// the singular direction of travel for a pawn
	d := PieceDelta[Pawn][pos.Turn]
	x := tile + d

	// advance forward once (can't be off board)
	if pos.Board[x] == nil {
		moves = append(moves, Move{
			Origin: tile,
			Dest: x,
			Pawn: true,
			Promote: Rank(x) == BackRank[pos.Turn.Opponent()],
		})

		// try pushing the pawn?
		if Rank(tile) == PawnRank[pos.Turn] {
			if pos.Board[x + d] == nil {
				moves = append(moves, Move{
					Origin: tile,
					Dest: x + d,
					Pawn: true,
					Push: true,
				})
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
			moves = append(moves, Move{
				Origin: tile,
				Dest: x + i,
				Pawn: true,
				Capture: true,
				EnPassant: true,
			})
		} else {
			p := pos.Board[x + i]

			if p != nil && p.Color != pos.Turn {
				moves = append(moves, Move{
					Origin: tile,
					Dest: x + i,
					Pawn: true,
					Capture: true,
					Promote: Rank(x + i) == BackRank[pos.Turn.Opponent()],
				})
			}
		}
	}

	return moves
}

func (pos *Position) NonPawnMoves(tile int, kind Kind) []Move {
	moves := make([]Move, 0, 28)

	// traverse in every direction the piece can move
	for _, d := range PieceDelta[kind] {
		capture := false

		// sliding pieces keep moving along that direction
		for x := tile + d; capture || Offboard(x) == false; x += d {
			if p := pos.Board[x]; p != nil {
				if p.Color != pos.Turn {
					capture = true
				} else {
					break
				}
			}

			moves = append(moves, Move{
				Origin: tile,
				Dest: x,
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

func (pos *Position) CastleMoves() []Move {
	tile := Tile(BackRank[pos.Turn], 4)
	moves := make([]Move, 0, 2)

	// is the kingside castle available?
	if pos.Castles & (Kingside << uint(pos.Turn << 2)) != 0 {
		b := pos.Board[tile + 1] == nil
		n := pos.Board[tile + 2] == nil

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
	if pos.Castles & (Queenside << uint(pos.Turn << 2)) != 0 {
		q := pos.Board[tile - 1] == nil
		b := pos.Board[tile - 2] == nil
		n := pos.Board[tile - 3] == nil

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
