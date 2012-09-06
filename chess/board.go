package chess

type Board struct {
	Position [128]*Piece   // 0x88 representation
}

var BackRank = [2]int{ 0, 7 }
var PawnRank = [2]int{ 1, 6 }

func Tile(rank, file int) int {
	return rank << 4 + file
}

func (b *Board) Setup() {
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
	if tile & 0x88 == 0 {
		b.Position[tile] = &Piece{Color: color, Kind: kind}
	}
}

func (b *Board) Move(origin, dest int) {
	if (origin | dest) & 0x88 == 0 {
		b.Position[dest] = b.Position[origin]
		b.Position[origin] = nil
	}
}

func (b *Board) Print() {
	fmt.Println("  +---+---+---+---+---+---+---+---+")

	for rank := 7; rank >= 0; rank-- {
		fmt.Printf("%d |", rank + 1)

		// rank axis
		for file := 0; file < 8; file++ {
			fmt.Printf(" %c |", b.Position[Tile(rank, file)].Rune())
		}

		fmt.Println("\n  +---+---+---+---+---+---+---+---+")
	}

	// file axis
	fmt.Println("    a   b   c   d   e   f   g   h")
}
