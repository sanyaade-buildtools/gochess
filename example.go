package main

import (
	"fmt"
)

import (
	"./chess"
	"./fen"
//	"./pgn"
)

func main() {
	pos := fen.Parse(fen.Start)

	pos.Board.Print()

	moves := chess.Eval(pos, chess.White)
	fmt.Println(len(moves))
}
