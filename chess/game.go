package chess

type Game struct {
	Board Board               // 0x88 representation
	Turn Color                // white or black
	Attacks [2]map[int][]int  // what tiles are attacked
	King [2]int               // location of each player's king
	Castle int                // what castle moves are available
	EnPassant int             // tile en passant can be performed into
	HalfMove int              // pawn half moves
	Moves []Move              // all moves made
}

type Move struct {
	Origin, Dest int          // where it is moving from and to
	Capture bool              // captured another piece
	Castle int                // castle move: Kingside or Queenside
	EnPassant bool            // was an en passant capture
	Pawn, Push, Promote bool  // pawn move, 2 space push, promotion
	Kind Kind                 // what was moved or promotion
}

// Game.Castle is a 4-bit nibble, where each 2 bits represents
// the kingside and queenside availability. To get the bit value
// for a given player = Side << (Color << 2)

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
	g.King[White] = Tile(0, 4)
	g.King[Black] = Tile(7, 4)
	g.Castle = 15
	g.EnPassant = -1
	g.HalfMove = 0
	g.Moves = make([]Move, 0, 50)
}

func (g *Game) PerformMove(move Move) {
	if move.Castle != 0 {
		g.PerformCastle(move.Castle)
	} else {
		g.Board.Move(move.Origin, move.Dest)

		// pawn moves are special
		if move.Pawn {
			enPassant := move.Dest + PieceDelta[Pawn][g.Turn.Opponent()]

			switch {
				case move.EnPassant:
					g.Board.Remove(enPassant)
					break
				case move.Push:
					g.EnPassant = enPassant
					break
				case move.Promote:
					g.Board.Place(move.Dest, g.Turn, move.Kind)
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

	// update the king's position if moved
	if move.Kind == King {
		g.King[g.Turn] = move.Dest
		g.DisableCastle(Kingside | Queenside)
	}

	// record the move and switch whose turn it is
	g.Moves = append(g.Moves, move)
	g.Turn = g.Turn.Opponent()

	// update the half move counter
	if move.Pawn == true {
		g.HalfMove = 0
	} else {
		g.HalfMove++
	}
}

func (g *Game) PerformCastle(castle int) {
	rank := BackRank[g.Turn]

	switch castle {
		case Kingside:
			g.Board.Move(Tile(rank, 4), Tile(rank, 6))
			g.Board.Move(Tile(rank, 7), Tile(rank, 5))
			break
		case Queenside:
			g.Board.Move(Tile(rank, 4), Tile(rank, 2))
			g.Board.Move(Tile(rank, 0), Tile(rank, 3))
			break
	}

	g.DisableCastle(Kingside | Queenside)
}

func (g *Game) DisableCastle(side int) {
	g.Castle &= ^(side << uint(g.Turn << 2))
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

func (g *Game) ContainsWinningMove(moves []Move) bool {
	for _, move := range moves {
		if move.Capture && g.Board[move.Dest].Kind == King {
			return true
		}
	}
	return false
}
