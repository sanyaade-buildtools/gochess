package main

import (
	"fmt"
	"./chess"
//	"./fen"
//	"./pgn"
)

var moves = []string{
	"d3", "Nf6",
	"Nd2", "d5",
	"b3", "e5",
	"Bb2", "Nc6",
	"g3", "Bg4",
	"Bg2", "Bb4",
	"a3", "Bxd2+",
	"Qxd2", "O-O",
}

func main() {
	g := chess.NewGame()

	for _, move := range moves {
		x := g.ParseMove(move)
		fmt.Println(move, x)
		g.PerformMove(x)
		g.Position.Render()
	}
}
