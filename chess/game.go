package chess

import "fmt"

const (
	Kingside = 1 + iota
	Queenside
)

type Setup struct {
	Turn Color                // whose turn it is
	King [2]int               // location of king pieces
	EnPassant int             // en passant availability
	Castles int               // castling availability
	HalfMove int              // pawn half moves
	Move int                  // current full move
}

type Game struct {
	Position Board            // 0x88 board representation
	Setup Setup               // current game state
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
	Setup Setup               // game state copy
	Capture *Piece            // destination piece (if captured)
	Move *Move                // the move that was performed
}

func (g *Game) New() {
	g.Position.New()

	// initial state for a new game
	g.Setup.Turn = White
	g.Setup.King[White] = Tile(0, 4)
	g.Setup.King[Black] = Tile(7, 4)
	g.Setup.EnPassant = -1
	g.Setup.Castles = 15
	g.Setup.HalfMove = 0
	g.Setup.Move = 1
}

func (g *Game) Moves() []*Move {
	moves := make([]*Move, 0, 30)
	c := make(chan *Move)

	go g.PseudoLegalMoves(c)

	for move := range c {
		if g.IsLegalMove(move) {
			moves = append(moves, move)
		}
	}

	return moves
}

func (g *Game) IsLegalMove(move *Move) bool {
	if p := g.Position[move.Origin]; p == nil || p.Color != g.Setup.Turn {
		return false
	}

	// try making the move, which will update the position
	undo := g.PerformMove(move)

	// always make sure to undo the move
	defer g.PerformUndo(undo)

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
	return true //g.IsAttacked(g.Setup.King[g.Setup.Turn]) == false
}

func (g *Game) IsAttacked(tile int) bool {
	piece := g.Position[tile]
	t := piece.Color
	opp := t.Opponent()

	// check pawn attacks to the left
	if l := tile - PieceDelta[Pawn][opp] - 1; Offboard(l) == false {
		if p := g.Position[l]; p != nil && p.Color == opp {
			fmt.Println("a")
			return true
		}
	}

	// check pawn attack to the right
	if r := tile - PieceDelta[Pawn][opp] + 1; Offboard(r) == false {
		if p := g.Position[r]; p != nil && p.Color == opp {
			fmt.Println("b")
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
			fmt.Println("c")
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

			fmt.Println("d")
	return false
}

func (g *Game) PseudoLegalMoves(ch chan *Move) {
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			tile := Tile(rank, file)

			// find moves for pieces matching the color
			if p := g.Position[tile]; p != nil && p.Color == g.Setup.Turn {
				if p.Kind == Pawn {
					g.PawnMoves(ch, tile)
				} else {
					g.NonPawnMoves(ch, tile, p.Kind)
				}
			}
		}
	}

	// get available castle moves
	g.CastleMoves(ch)

	// done collecting moves
	close(ch)
}

func (g *Game) PawnMoves(ch chan *Move, tile int) {
	d := PieceDelta[Pawn][g.Setup.Turn]
	back := BackRank[g.Setup.Turn.Opponent()]
	x := tile + d

	// advance forward once (can't be off board)
	if g.Position[x] == nil {
		ch <- &Move{
			Origin: tile,
			Dest: x,
			Pawn: true,
			Promote: Rank(x) == back,
		}

		// try pushing the pawn?
		if Rank(tile) == PawnRank[g.Setup.Turn] {
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
		if tile + d + i == g.Setup.EnPassant {
			ch <- &Move{
				Origin: tile,
				Dest: x + i,
				Pawn: true,
				Capture: true,
				EnPassant: true,
			}
		} else {
			p := g.Position[x + i]

			if p != nil && p.Color != g.Setup.Turn {
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

func (g *Game) NonPawnMoves(ch chan *Move, tile int, kind Kind) {
	for _, d := range PieceDelta[kind] {
		capture := false

		// sliding pieces keep moving along that direction
		for x := tile + d; capture || Offboard(x) == false; x += d {
			if p := g.Position[x]; p != nil {
				if p.Color != g.Setup.Turn {
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
	tile := Tile(BackRank[g.Setup.Turn], 4)

	// is the kingside castle available?
	if g.Setup.Castles & (Kingside << uint(g.Setup.Turn << 2)) != 0 {
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
	if g.Setup.Castles & (Queenside << uint(g.Setup.Turn << 2)) != 0 {
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
func (g *Game) PerformMove(move *Move) *Undo {
	var capture *Piece

	// en passant is unavailable after every move
	g.Setup.EnPassant = -1

	// check for a castle move
	if move.Castle != 0 {
		rank := BackRank[g.Setup.Turn]

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
		if move.Capture {
			capture = g.Position[move.Dest]
		}

		// move the piece to the new position
		g.Position.Move(move.Origin, move.Dest)

		// pawn moves are special
		if move.Pawn {
			enPassant := move.Dest + PieceDelta[Pawn][g.Setup.Turn.Opponent()]

			switch {
				case move.EnPassant:
					g.Position.Remove(enPassant)
					break
				case move.Push:
					g.Setup.EnPassant = enPassant
					break
				case move.Promote:
					g.Position.Place(move.Dest, g.Setup.Turn, move.Kind)
					break
			}
		}

		// moving the rooks disables castling
		switch move.Origin {
			case Tile(BackRank[g.Setup.Turn], 0):
				g.DisableCastle(Queenside)
				break
			case Tile(BackRank[g.Setup.Turn], 7):
				g.DisableCastle(Kingside)
				break
		}
	}

	// update the king's position if moved
	if move.Kind == King {
		g.Setup.King[g.Setup.Turn] = move.Dest

		// moving the king disabled all castling
		g.DisableCastle(Kingside | Queenside)
	}

	// record the move and switch whose turn it is
	if g.Setup.Turn = g.Setup.Turn.Opponent(); g.Setup.Turn == White {
		g.Setup.Move++
	}

	// update the half move counter
	if move.Pawn == true {
		g.Setup.HalfMove = 0
	} else {
		g.Setup.HalfMove++
	}

	return &Undo{
		Setup: g.Setup,
		Capture: capture,
		Move: move,
	}
}

func (g *Game) DisableCastle(side int) {
	g.Setup.Castles &= ^(side << uint(g.Setup.Turn << 2))
}

func (g *Game) PerformUndo(undo *Undo) {
	if undo.Move.Castle != 0 {
		// TODO:
	} else {
		g.Position.Move(undo.Move.Dest, undo.Move.Origin)
		g.Position[undo.Move.Dest] = undo.Capture
	}

	// restore to the previous state
	g.Setup = undo.Setup
}
