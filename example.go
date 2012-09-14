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
	g := fen.Parse(fen.Start)

	g.Position.Render()

	moves := g.CollectMoves()

	for _, move := range moves {
		fmt.Println(move.LongNotation())
	}
}
