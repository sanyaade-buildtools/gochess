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

	pseudoMoves := g.PseudoLegalMoves()
	moves := g.LegalMoves(pseudoMoves)

	fmt.Println(len(moves))
}
