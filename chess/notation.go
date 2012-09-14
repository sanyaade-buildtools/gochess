package chess

import "fmt"

func TileNotation(tile int) string {
	return fmt.Sprintf("%c%d", byte('a') + byte(File(tile)), 1 + Rank(tile))
}

func (move *Move) LongNotation() string {
	switch move.Castle {
		case Kingside: return "O-O"
		case Queenside: return "O-O-O"
	}

	piece := PieceRunes[White][move.Kind]
	origin := TileNotation(move.Origin)
	dest := TileNotation(move.Dest)
	x := '-'

	if move.Capture {
		x = 'x'
	}

	if move.Pawn {
		if move.Promote {
			return fmt.Sprintf("%s%c%s=%c", origin, x, dest, piece)
		} else {
			return fmt.Sprintf("%s%c%s", origin, x, dest)
		}
	}

	return fmt.Sprintf("%c%s%c%s", piece, origin, x, dest)
}
