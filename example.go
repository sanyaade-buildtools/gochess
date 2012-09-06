package main

import (
	"./chess"
	"fmt"
)

func main() {
	b := chess.NewBoard()

	chess.PrintBoard(b)
	fmt.Println(b.ValidMoves())
}
