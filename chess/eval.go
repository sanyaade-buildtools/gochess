package chess

import (
	"fmt"
	"regexp"
	"unicode"
)

// regular expression for parsing short-/long-hand algebraic moves
var reMove = regexp.MustCompile(
	"^O-(?:O-)?O|([PNBRQK])?([a-h]?[1-8]?)(x|-)?([a-h][1-8])(=[NBRQ])?[+#]?$",
)

var PawnAttackTable = [2][]int{
	[]int{ -15, -17 },
	[]int{ 15, 17 },
}

const (
	Attack_BQ = (1 << uint(Bishop)) | (1 << uint(Queen))
	Attack_RQ = (1 << uint(Rook)) | (1 << uint(Queen))
	Attack_N  = (1 << uint(Knight))
	Attack_K  = (1 << uint(King))
)

var AttackTable = map[int][][]int{
	Attack_BQ: [][]int{
		[]int{ 15, 30, 45, 60, 75, 90, 105 }, // up left
		[]int{ 17, 34, 51, 68, 85, 102, 119 }, // up right
		[]int{ -15, -30, -45, -60, -75, -90, -105 }, // down left
		[]int{ -17, -34, -51, -68, -85, -102, -119 }, // down right
	},
	Attack_RQ: [][]int{
		[]int{ 16, 32, 48, 64, 80, 96, 112 }, // up
		[]int{ 1, 2, 3, 4, 5, 6, 7 }, // right
		[]int{ -16, -32, -48, -64, -80, -96, -112 }, // down
		[]int{ -1, -2, -3, -4, -5, -6, -7 }, // left
	},
	Attack_N: [][]int{ PieceDelta[Knight] },
	Attack_K: [][]int{ PieceDelta[King] },
}

func (g *Game) IsLegalMove(move *Move) bool {
	p := g.Position.Piece(move.Origin)
	x := g.Position.Piece(move.Dest)

	if Offboard(move.Dest) || p == nil {
		return false
	}

	// the piece exists and is owned by the current player
	if p == nil || p.Color != g.Turn {
		return false
	}

	if move.Castle != 0 {
		if g.Castles & (move.Castle << uint(g.Turn << 2)) == 0 {
			return false
		}

		// queenside castles move down in file, kingside up
		if move.Castle == Kingside {
			for i := 0; i < 3; i++ {
				if g.InCheck(g.King[g.Turn] + i) {
					return false
				}
			}
		} else {
			for i := 0; i < 3; i++ {
				if g.InCheck(g.King[g.Turn] - i) {
					return false
				}
			}
		}

		return true
	}

	if move.Pawn {
		if p.Kind != Pawn {
			return false
		}

		// can only push pawns on the pawn rank
		if move.Push && Rank(move.Origin) != PawnRank[g.Turn] {
			return false
		}

		// make sure en passant is valid
		if move.EnPassant && move.Dest != g.EnPassant {
			return false
		}
	}

	// undo regardless
	defer func() {
		g.Position[move.Origin] = p
		g.Position[move.Dest] = x
	}()

	// make move
	g.Position[move.Origin] = nil
	g.Position[move.Dest] = p

	// get the king's location
	king := g.King[g.Turn]

	// if the king moved, update the king's location
	if move.Kind == King {
		king = move.Dest
	}

	return g.InCheck(king) == false
}

func (g *Game) InCheck(tile int) bool {
	opp := g.Turn.Opponent()

	// check for simple pawn attacks
	for delta := range PawnAttackTable[opp] {
		if p := g.Position.Piece(tile + delta); p != nil {
			if p.Color == g.Turn.Opponent() && p.Kind == Pawn {
				return true
			}
		}
	}

	// non-pawn attacking pieces
	for pieces, attacks := range AttackTable {
		for _, direction := range attacks {
			for _, delta := range direction {
				if p := g.Position.Piece(tile + delta); p != nil {
					if pieces & (1 << uint(p.Kind)) != 0 {
						if p.Color == opp {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

func (g *Game) CollectMoves() []*Move {
	pseudoMoves := make(chan *Move)
	moves := make([]*Move, 0, 30)

	go func() {
		for rank := 0; rank < 8; rank++ {
			for file := 0; file < 8; file++ {
				tile := Tile(rank, file)

				// only collect moves for this player
				if p := g.Position[tile]; p != nil && p.Color == g.Turn {
					if p.Kind == Pawn {
						g.PawnMoves(pseudoMoves, tile)
					} else {
						g.NonPawnMoves(pseudoMoves, tile, p.Kind)
					}
				}
			}
		}

		// add castling moves
		g.CastleMoves(pseudoMoves)

		// all pseudo legal moves have been collected
		close(pseudoMoves)
	}()

	// filter legal moves from the pseudo legal ones
	for move := range pseudoMoves {
		if g.IsLegalMove(move) {
			moves = append(moves, move)
		}
	}

	return moves
}

func (g *Game) PawnMoves(ch chan *Move, tile int) {
	d := PieceDelta[Pawn][g.Turn]
	x := tile + d

	// advance forward once (can't be off board)
	if g.Position[x] == nil {
		ch <- &Move{
			Origin: tile,
			Dest: x,
			Pawn: true,
			Promote: Rank(x) == BackRank[g.Turn.Opponent()],
		}

		// try pushing the pawn?
		if Rank(tile) == PawnRank[g.Turn] {
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
		if tile + d + i == g.EnPassant {
			ch <- &Move{
				Origin: tile,
				Dest: x + i,
				Pawn: true,
				Capture: true,
				EnPassant: true,
			}
		} else {
			p := g.Position[x + i]

			if p != nil && p.Color != g.Turn {
				ch <- &Move{
					Origin: tile,
					Dest: x + i,
					Pawn: true,
					Capture: true,
					Promote: Rank(x + i) == BackRank[g.Turn.Opponent()],
				}
			}
		}
	}
}

func (g *Game) NonPawnMoves(ch chan *Move, tile int, kind Kind) {
	for _, d := range PieceDelta[kind] {
		capture := false

		// sliding pieces keep moving along that direction
		for x := tile + d; !Offboard(x) && !capture; x += d {
			if p := g.Position.Piece(x); p != nil {
				if p.Color != g.Turn {
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
	tile := Tile(BackRank[g.Turn], 4)

	// is the kingside castle available?
	if g.Castles & (Kingside << uint(g.Turn << 2)) != 0 {
		b := g.Position.Piece(tile + 1) == nil
		n := g.Position.Piece(tile + 2) == nil

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
	if g.Castles & (Queenside << uint(g.Turn << 2)) != 0 {
		q := g.Position.Piece(tile - 1) == nil
		b := g.Position.Piece(tile - 2) == nil
		n := g.Position.Piece(tile - 3) == nil

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

func (g *Game) ParseMove(s string) *Move {
	var m []string
	var castle int
	var k Kind
	var move *Move

	// try and parse the move string
	if m = reMove.FindStringSubmatch(s); m == nil {
		fmt.Println("HERE")
		return nil
	}

	// get all the available moves
	moves := g.CollectMoves()

	fmt.Println(moves)

	// check for a castling move
	switch m[0] {
		case "O-O":   castle = Kingside; break
		case "O-O-O": castle = Queenside; break
	}

	// get the piece kind being moved
	switch m[1] {
		case "P", "": k = Pawn; break
		case "N":     k = Knight; break
		case "B":     k = Bishop; break
		case "R":     k = Rook; break
		case "Q":     k = Queen; break
		case "K":     k = King; break
	}

	// determine if there has been a capture
	x := m[3] == "x"

	// parse the final, destination tile being moved to
	tile := Tile(
		int(byte(m[4][1]) - byte('1')),
		int(byte(m[4][0]) - byte('a')),
	)

	// check for a pawn promotion
	promote := len(m[5]) > 0

	// determine if this move can match
	filter := func(move *Move) bool {
		switch {
			case move.Castle != castle:   return false
			case move.Dest != tile:       return false
			case move.Capture != x:       return false
			case move.Promote != promote: return false
		}

		// pawn move or same piece being moved
		return move.Pawn && k == Pawn || move.Kind == k
	}

	// filter moves targeting the same destination and capture
	for i := 0; i < len(moves); {
		if filter(moves[i]) == false {
			moves[i] = moves[len(moves) - 1]
			moves = moves[:len(moves) - 1]
		} else {
			i++
		}
	}

	// no matching legal moves left?
	if len(moves) == 0 {
		fmt.Println("HERE2")
		return nil
	}

	// move origin
	rank := -1
	file := -1

	// the origin can be the file, rank, or both
	switch len(m[2]) {
		case 1:
			if unicode.IsDigit(rune(m[2][0])) {
				rank = int(byte(m[2][0]) - byte('1'))
			} else {
				file = int(byte(m[2][0]) - byte('a'))
			}
			break
		case 2:
			file = int(byte(m[2][0]) - byte('a'))
			rank = int(byte(m[2][1]) - byte('1'))
			break
	}

	// find the move that best matches the origin
	for i := 0; i < len(moves); i++ {
		if file >= 0 && File(moves[i].Origin) != file { continue }
		if rank >= 0 && Rank(moves[i].Origin) != rank { continue }

		// the rank or the file matches
		if move != nil {
			fmt.Println("HERE3")
			return nil
		}

		// this move matches the origin
		move = moves[i]
	}

	// set the promotion kind for the move
	if move.Promote && promote {
		move.Kind = k
	}

	return move
}
