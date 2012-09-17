package main

import (
	"./chess"
//	"./fen"
//	"./pgn"
)

func main() {
	g := chess.NewGame()

	g.PerformMove(g.ParseMove("e4"))
	g.PerformMove(g.ParseMove("e5"))

	g.Position.Render()
}
