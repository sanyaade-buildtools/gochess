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

func Parse(fen string) *chess.Position {
	pos := new(chess.Position)

	// initialize an empty position
	pos.Init()

	// divide the FEN into its components
	sections := strings.Split(fen, " ")

	if len(sections) != 6 {
		return nil
	}

	// initialize each part of the game
	if !setBoard(pos, sections[0]) { return nil }
	if !setTurn(pos, sections[1]) { return nil }
	if !setCastle(pos, sections[2]) { return nil }
	if !setEnPassant(pos, sections[3]) { return nil }
	if !setHalfMove(pos, sections[4]) { return nil }
	if !setMove(pos, sections[5]) { return nil }

	return pos
}

func setBoard(pos *chess.Position, setup string) bool {
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
				pos.Board.Place(tile, p.Color, p.Kind)

				// save the king locations
				if p.Kind == chess.King {
					pos.King[p.Color] = tile
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

func setTurn(pos *chess.Position, turn string) bool {
	switch turn {
		case "w", "W": pos.Turn = chess.White; return true
		case "b", "B": pos.Turn = chess.Black; return true
	}

	return false
}

func setCastle(pos *chess.Position, castle string) bool {
	if castle == "-" {
		return true
	}

	for _, c := range castle {
		switch c {
			case 'K': 
				pos.Castles |= chess.Kingside << uint(chess.White << 2)
				break
			case 'Q':
				pos.Castles |= chess.Queenside << uint(chess.White << 2)
				break
			case 'k':
				pos.Castles |= chess.Kingside << uint(chess.Black << 2)
				break
			case 'q':
				pos.Castles |= chess.Queenside << uint(chess.Black << 2)
				break

			default:
				return false
		}
	}

	return true
}

func setEnPassant(pos *chess.Position, ep string) bool {
	if ep == "-" {
		pos.EnPassant = -1
	} else {
		if len(ep) != 2 {
			return false
		}

		file := int(ep[0]) - int('a')
		rank := int(ep[1]) - int('1')

		if file < 0 || file > 7 || (rank != 1 && rank != 6) {
			return false
		}

		pos.EnPassant = chess.Tile(rank, file)
	}

	return true
}

func setHalfMove(pos *chess.Position, half string) bool {
	n, err := strconv.Atoi(half)

	if err == nil {
		pos.HalfMove = n
	}

	return err == nil
}

func setMove(pos *chess.Position, move string) bool {
	n, err := strconv.Atoi(move)

	if err == nil {
		pos.Start = n
	}

	return err == nil
}
