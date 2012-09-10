package chess

const (
	Kingside = 1 + iota
	Queenside
)

type Game struct {
	Turn Color                // whose turn it is
	EnPassant int             // en passant availability
	Castles int               // castling availability
	HalfMove int              // pawn half moves
	Move int                  // current full move
}

type Position struct {
	Game Game                 // current game state
	Board Board               // 0x88 board representation
	King [2]int               // location of the kings
}

type Move struct {
	Origin, Dest int          // where it is moving from and to
	Capture bool              // captured another piece
	Castle int                // castle move: Kingside or Queenside
	EnPassant bool            // was an en passant capture
	Pawn, Push, Promote bool  // pawn move, 2 space push, promotion
	Kind Kind                 // what was moved or promotion
}

type Undo struct {
	Game Game                 // game state copy
	Piece *Piece              // destination piece (if captured)
	Move *Move                // the move that was performed
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

func (pos *Position) Eval() []*Move {
	moves := make([]*Move, 0, 30)
	c := make(chan *Move)

	go pos.PseudoLegalMoves(c)

	for move := range c {
		if pos.IsLegalMove(move) {
			moves = append(moves, move)
		}
	}

	return moves
}

func (pos *Position) IsLegalMove(move *Move) bool {
	turn := pos.Turn
	undo := pos.PerformMove(move)

	// always make sure to undo the move
	defer pos.PerformUndo(undo)

	switch move.Castle {
		case Kingside:
			for x := move.Origin; x < move.Dest; x++ {
				if pos.IsAttacked(x) { return false }
			}
			break
		case Queenside:
			for x := move.Origin; x > move.Dest; x-- {
				if pos.IsAttacked(x) { return false }
			}
			break
	}

	return pos.IsAttacked(pos.King[turn]) == false
}

func (pos *Position) IsAttacked(tile int) bool {
	piece := pos.Board[tile]
	t := piece.Color
	opp := t.Opponent()

	// check pawn attacks to the left
	if l := tile + PieceDelta[Pawn][opp] - 1; Offboard(l) == false {
		if p := pos.Board[l]; p != nil && p.Color != t {
			return true
		}
	}

	// check pawn attack to the right
	if r := tile + PieceDelta[Pawn][opp] + 1; Offboard(r) == false {
		if p := pos.Board[r]; p != nil && p.Color != t {
			return true
		}
	}

	// check non-pawn piece types
	for kind := 1; kind < 6; kind++ {
		for _, d := range PieceDelta[kind] {
			capture := false

			// sliding pieces keep moving along that direction
			for x := tile + d; capture || Offboard(x) == false; x += d {
				if p := pos.Board[x]; p != nil {
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

func (pos *Position) PseudoLegalMoves(ch chan *Move) {
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			tile := Tile(rank, file)

			// find moves for pieces matching the color
			if p := pos.Board[tile]; p != nil && p.Color == pos.Turn {
				if p.Kind == Pawn {
					pos.PawnMoves(ch, tile)
				} else {
					pos.NonPawnMoves(ch, tile, p.Kind)
				}
			}
		}
	}

	// get available castle moves
	pos.CastleMoves(ch)

	// done collecting moves
	close(ch)
}

func (pos *Position) PawnMoves(ch chan *Move, tile int) {
	d := PieceDelta[Pawn][pos.Turn]
	back := BackRank[pos.Turn.Opponent()]
	x := tile + d

	// advance forward once (can't be off board)
	if pos.Board[x] == nil {
		ch <- &Move{
			Origin: tile,
			Dest: x,
			Pawn: true,
			Promote: Rank(x) == back,
		}

		// try pushing the pawn?
		if Rank(tile) == PawnRank[pos.Turn] {
			if pos.Board[x + d] == nil {
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
		if tile + d + i == pos.EnPassant {
			ch <- &Move{
				Origin: tile,
				Dest: x + i,
				Pawn: true,
				Capture: true,
				EnPassant: true,
			}
		} else {
			p := pos.Board[x + i]

			if p != nil && p.Color != pos.Turn {
				ch <- &Move{
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

func (pos *Position) NonPawnMoves(ch chan *Move, tile int, kind Kind) {
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

func (pos *Position) CastleMoves(ch chan *Move) {
	tile := Tile(BackRank[pos.Turn], 4)

	// is the kingside castle available?
	if pos.Castles & (Kingside << uint(pos.Turn << 2)) != 0 {
		b := pos.Board[tile + 1] == nil
		n := pos.Board[tile + 2] == nil

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
	if pos.Castles & (Queenside << uint(pos.Turn << 2)) != 0 {
		q := pos.Board[tile - 1] == nil
		b := pos.Board[tile - 2] == nil
		n := pos.Board[tile - 3] == nil

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
func (pos *Position) PerformMove(move *Move) *Undo {
	undo := &Undo{}

	if move.Castle != 0 {
		rank := BackRank[pos.Turn]

		switch move.Castle {
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

	return undo
}

func (pos *Position) PerformCastle(castle int) {
}

func (pos *Position) DisableCastle(side int) {
	pos.Castles &= ^(side << uint(pos.Turn << 2))
}

func (pos *Position) PerformUndo(undo *Undo) {
	if undo.Move.Castle != 0 {
		// TODO:
	} else {
		if undo.Move.Promote {
			pos.Board.Place(undo.Move.Origin, undo.Turn, Pawn)
		} else {
			pos.Board.Move(undo.Move.Dest, undo.Move.Origin)
			pos.Board[undo.Move.Dest] = undo.Piece
		}
	}
}
