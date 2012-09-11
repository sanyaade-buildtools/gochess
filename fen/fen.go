package fen

import (
	"strings"
	"strconv"
)

import "../chess"

const (
	Start = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
)

var PieceMap = map[rune]chess.Piece{
	'P': chess.Piece{Color: chess.White, chess.Kind: chess.Pawn},
	'B': chess.Piece{Color: chess.White, chess.Kind: chess.Bishop},
	'N': chess.Piece{Color: chess.White, chess.Kind: chess.Knight},
	'R': chess.Piece{Color: chess.White, chess.Kind: chess.Rook},
	'Q': chess.Piece{Color: chess.White, chess.Kind: chess.Queen},
	'K': chess.Piece{Color: chess.White, chess.Kind: chess.King},
	'p': chess.Piece{Color: chess.Black, chess.Kind: chess.Pawn},
	'b': chess.Piece{Color: chess.Black, chess.Kind: chess.Bishop},
	'n': chess.Piece{Color: chess.Black, chess.Kind: chess.Knight},
	'r': chess.Piece{Color: chess.Black, chess.Kind: chess.Rook},
	'q': chess.Piece{Color: chess.Black, chess.Kind: chess.Queen},
	'k': chess.Piece{Color: chess.Black, chess.Kind: chess.King},
}

func Parse(fen string) *chess.Game {
	g := new(chess.Game)

	// divide the FEN into its components
	sections := strings.Split(fen, " ")

	if len(sections) != 6 {
		return nil
	}

	// initialize each part of the game
	if !setBoard(g, sections[0]) { return nil }
	if !setTurn(g, sections[1]) { return nil }
	if !setCastle(g, sections[2]) { return nil }
	if !setEnPassant(g, sections[3]) { return nil }
	if !setHalfMove(g, sections[4]) { return nil }
	if !setMove(g, sections[5]) { return nil }

	return g
}

func setBoard(g *chess.Game, setup string) bool {
	ranks := strings.Split(setup, "/")

	// make sure there were 8 ranks of data
	if len(ranks) != 8 {
		return false
	}

	// loop over all the ranks, starting from black's side
	for rank := 7; rank >= 0; rank-- {
		file := 0

		for _, c := range ranks[7 - rank] {
			if c >= '1' && c <= '8' {
				file += int(c) - int('0')
			} else {
				tile := chess.Tile(rank, file)
				p, ok := PieceMap[c]

				if ok == false {
					return false
				}

				// put the piece onto the board
				g.Position.Place(tile, p.Color, p.Kind)

				// save the king locations
				if p.Kind == chess.King {
					g.Setup.King[p.Color] = tile
				}

				// advance
				file++
			}
		}

		// make sure the entire rank was set
		if file != 8 {
			return false
		}
	}

	return true
}

func setTurn(g *chess.Game, turn string) bool {
	switch turn {
		case "w", "W": g.Setup.Turn = chess.White; return true
		case "b", "B": g.Setup.Turn = chess.Black; return true
	}

	return false
}

func setCastle(g *chess.Game, castle string) bool {
	if castle == "-" {
		return true
	}

	for _, c := range castle {
		switch c {
			case 'K': 
				g.Setup.Castles |= chess.Kingside << uint(chess.White << 2)
				break
			case 'Q':
				g.Setup.Castles |= chess.Queenside << uint(chess.White << 2)
				break
			case 'k':
				g.Setup.Castles |= chess.Kingside << uint(chess.Black << 2)
				break
			case 'q':
				g.Setup.Castles |= chess.Queenside << uint(chess.Black << 2)
				break

			default:
				return false
		}
	}

	return true
}

func setEnPassant(g *chess.Game, ep string) bool {
	if ep == "-" {
		g.Setup.EnPassant = -1
	} else {
		if len(ep) != 2 {
			return false
		}

		file := int(ep[0]) - int('a')
		rank := int(ep[1]) - int('1')

		if file < 0 || file > 7 || (rank != 1 && rank != 6) {
			return false
		}

		g.Setup.EnPassant = chess.Tile(rank, file)
	}

	return true
}

func setHalfMove(g *chess.Game, half string) bool {
	n, err := strconv.Atoi(half)

	if err == nil {
		g.Setup.HalfMove = n
	}

	return err == nil
}

func setMove(g *chess.Game, move string) bool {
	n, err := strconv.Atoi(move)

	if err == nil {
		g.Setup.Move = n
	}

	return err == nil
}
