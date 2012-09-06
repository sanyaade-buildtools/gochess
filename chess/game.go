package chess

type Game struct {
	Board Board               // position representation
	Turn Color                // white or black
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


