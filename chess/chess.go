package chess

import (
	"fmt"
)

const (
	Kingside = 1 + iota
	Queenside
)

type Game struct {
	Position [128]*Piece      // 0x88 board representation
	Turn Color                // White or Black
	Castling int              // what castle moves are available
	EnPassant int             // tile en passant can be performed into
	HalfMove, Move int        // pawn half moves and full move count
}

type Move struct {
	Origin, Dest int          // where it is moving from and to
	Capture bool              // captured another piece
	Castle int                // castle move (0=not a castle)
	EnPassant bool            // was an en passant capture
	Pawn, Push, Promote bool  // pawn move, 2 space push, promotion
	Kind Kind                 // promote piece, ...
}


var PieceDelta = [6][]int{
	[]int{ },
	[]int{ -15, 15, -17, 17 },
	[]int{ -31, -33, -14, 18, -18, 14, 31, 33 },
	[]int{ -1, 1, -16, 16, -17, 15, -15, 17 },
	[]int{ -1, 1, -16, 16 },
	[]int{ -1, 1, -16, 16, -17, 15, -15, 17 },
}


func (b *Board) PseudoLegalMoves(tile int, kind Kind) []Move {
	moves := make([]Move, 0, 32)

	for _, d := range PieceDelta[kind] {
		for pos := tile + d; ; pos += d {
			legal, capture := b.IsPseudoLegalMove(pos)

			if legal == false {
				break
			}

			moves = append(moves,Move{
				Origin: tile,
				Dest: pos,
				Capture: capture,
			})

			// stop if we captured something
			if capture == true {
				break
			}

			// non-sliding piece, don't keep advancing
			if kind & 1 == 0 {
				break
			}
		}
	}

	return moves
}

func (b *Board) IsPseudoLegalMove(tile int) (bool, bool) {
	if tile & 0x88 != 0 {
		return false, false
	}

	if piece := b.Position[tile]; piece != nil {
		return piece.Color != b.Turn, piece.Color != b.Turn
	}

	return true, false
}

func (b *Board) ValidMoves() []Move {
	var moves []Move

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			tile := Tile(rank, file)

			if piece := b.Position[tile]; piece != nil {
				if piece.Color == b.Turn {
					moves = append(moves, b.PseudoLegalMoves(tile, piece.Kind)...)
				}
			}
		}
	}

	return moves
}

func (b *Board) Perform(move *Move) bool {
	if (move.Origin | move.Dest) & 0x88 != 0 {
		return false
	}

	p := b.Position[move.Origin]
	//x := b.Position[move.Dest]

	// make sure the player is moving their piece
	if p.Color != b.Turn {
		return false
	}

	// handle en passant captures
	if move.EnPassant {
		if move.Dest != b.EnPassant {
			return false
		}

		// capture the pawn
		switch b.Turn {
			case White: b.Position[move.Dest - 16] = nil; break
			case Black: b.Position[move.Dest + 16] = nil; break
		}
	}

	// update the board
	b.Replace(move.Origin, move.Dest)

	// update half move
	if move.Pawn {
		b.HalfMove = 0
	} else {
		b.HalfMove++
	}

	// next player's turn, advance the move count
	if b.Turn = 1 - b.Turn; b.Turn == White {
		b.Move++
	}

	return true
}

// perform a castle on one side or the other
func (b *Board) Castle(side Color) {
	rank := BackRank[b.Turn]

	switch side {
		case Kingside:
			b.Replace(Tile(rank, 4), Tile(rank, 6))
			b.Replace(Tile(rank, 7), Tile(rank, 5))
			break
		case Queenside:
			b.Replace(Tile(rank, 4), Tile(rank, 2))
			b.Replace(Tile(rank, 0), Tile(rank, 3))
			break
	}

	// castling for this player is no longer allowed
	b.Castling &= ^(0x3 << uint(b.Turn << 1))
}

