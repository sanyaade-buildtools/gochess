package main

import (
//	"fmt"
)

import (
//	"./chess"
//	"./fen"
	"./pgn"
)

var game = "/Users/jeff/Dropbox/Public/Chess Games/AaronLedlie_vs_massung_2012_08_30.pgn"

func main() {
	pgn.Parse(game)
}
