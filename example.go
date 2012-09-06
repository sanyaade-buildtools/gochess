package main

import (
	"fmt"
)

import (
//	"./chess"
	"./fen"
)

func main() {
	g := fen.Parse(fen.Start)

	g.Board.Print()

	fmt.Println(g.PseudoLegalMoves())
}
