package main

import (
	"./chess"
	"fmt"
)

func main() {
	g := new(chess.Game)

	g.NewGame()
	g.Board.Print()

	fmt.Println(g.PseudoLegalMoves())
}
