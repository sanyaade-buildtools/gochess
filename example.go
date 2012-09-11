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

	moves := g.Moves()
	fmt.Println(len(moves))
}
