package main

import (
	"fmt"
)

import (
//	"./chess"
	"./fen"
//	"./pgn"
)

func main() {
	pos := fen.Parse(fen.Start)

	pos.Board.Print()

	fmt.Println(pos.PseudoLegalMoves())
}
