package chess

import "fmt"

type Board [128]*Piece

var BackRank = [2]int{ 0, 7 }
var PawnRank = [2]int{ 1, 6 }

func Tile(rank, file int) int {
	return rank << 4 + file
}

func Rank(tile int) int {
	return tile >> 4
}

func File(tile int) int {
	return tile & 7
}

func Offboard(tile int) bool {
	return tile & 0x88 != 0
}

func (b *Board) Clear() {
	for i := 0; i < 128; i++ {
		b[i] = nil
	}
}

func (b *Board) New() {
	backRow := [...]Kind{
		Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook,
	}

	for file := 0; file < 8; file++ {
		b.Place(Tile(BackRank[White], file), White, backRow[file])
		b.Place(Tile(BackRank[Black], file), Black, backRow[file])
		b.Place(Tile(PawnRank[White], file), White, Pawn)
		b.Place(Tile(PawnRank[Black], file), Black, Pawn)
	}
}

func (b *Board) Place(tile int, color Color, kind Kind) {
	if Offboard(tile) == false {
		b[tile] = &Piece{Color: color, Kind: kind}
	}
}

func (b *Board) Piece(tile int) *Piece {
	if Offboard(tile) == false {
		return b[tile]
	}
	return nil
}

func (b *Board) Move(origin, dest int) {
	if Offboard(origin | dest) == false {
		b[dest] = b[origin]
		b[origin] = nil
	}
}

func (b *Board) Remove(tile int) {
	if Offboard(tile) == false {
		b[tile] = nil
	}
}

func (b *Board) Render() {
	fmt.Println("  +---+---+---+---+---+---+---+---+")

	for rank := 7; rank >= 0; rank-- {
		fmt.Printf("%d |", rank + 1)

		// rank axis
		for file := 0; file < 8; file++ {
			fmt.Printf(" %c |", b[Tile(rank, file)].Rune())
		}

		fmt.Println("\n  +---+---+---+---+---+---+---+---+")
	}

	// file axis
	fmt.Println("    a   b   c   d   e   f   g   h")
}
