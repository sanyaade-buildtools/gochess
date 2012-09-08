package chess

type Color int
type Kind int

const (
	White Color = iota
	Black
)

const (
	Pawn Kind = iota
	Bishop
	Knight
	Rook
	King
	Queen
)

type Piece struct {
	Color Color
	Kind Kind
}

var PieceRunes = [2][6]rune{
	{ 'P', 'B', 'N', 'R', 'K', 'Q' },
	{ 'p', 'b', 'n', 'r', 'k', 'q' },
}

var PieceDelta = [6][]int{
	[]int{ 16, -16 },
	[]int{ -15, 15, -17, 17 },
	[]int{ -31, -33, -14, 18, -18, 14, 31, 33 },
	[]int{ -1, 1, -16, 16, -17, 15, -15, 17 },
	[]int{ -1, 1, -16, 16 },
	[]int{ -1, 1, -16, 16, -17, 15, -15, 17 },
}

func (p *Piece) Rune() rune {
	if p != nil {
		return PieceRunes[p.Color][p.Kind]
	}

	return ' '
}

func (k Kind) Sliding() bool {
	return int(k) & 1 != 0
}

func (c Color) Opponent() Color {
	return Black - c
}
