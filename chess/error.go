package chess

type Errno int

// errors when attempting invalid operations in a chess game
const (
	ParseError = Errno(iota)
	UnrecognizedPiece = Errno(iota)
	IllegalMove = Errno(iota)
	IllegalCastle = Errno(iota)
	AmbiguousMove = Errno(iota)
	InvalidPromotion = Errno(iota)
)

// error mappings
var errmap = map[Errno]string{
	ParseError: "Parse error",
	UnrecognizedPiece: "Unregognized piece",
	IllegalMove: "Illegal move",
	IllegalCastle: "Illegal castle",
	AmbiguousMove: "Ambiguous move",
	InvalidPromotion: "Invalid pawn promotion",
}

func (e Errno) Error() string {
	if msg, ok := errmap[e]; ok {
		return msg
	}
	return "Unknown error"
}
