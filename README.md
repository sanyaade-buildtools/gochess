# Go Chess Packages

This is a collection of tools for dealing with games of Chess!

# Installation

	go get github.com/massung/gochess

# Usage

There are 3 packages included with GoChess:

* chess
* fen
* pgn

These can all be imported independently, although the `chess` package is used by both the `fen` and `pgn` packages, so you should probably always import it.

	import (
		"github.com/massung/gochess/chess"
		"github.com/massung/gochess/fen"
		"github.com/massung/gochess/pgn"
	)

# The `chess` Package

The `chess` package is where most of your "chess" related work will actually be done. It is where the definitions for pieces, board positions, valid moves, castling, en passant, and more are.

## Creating a New Game

Use the `chess.Game` structure to hold any and all information about a game of chess. Simply call `NewGame()` one to get a brand new setup.

	g := new(chess.Game)
	g.NewGame()

## Displaying the Board

Once you have a `Game` structure, you can print out the `Position` of it at any time using the `Render()` function in the `Board`.

	g.Position.Board.Render()

This will output a `Board` to `os.Stdout`, which will print something like...

	  +---+---+---+---+---+---+---+---+
	8 | r | n | b | q | k | b | n | r |
	  +---+---+---+---+---+---+---+---+
	7 | p | p | p | p | p | p | p | p |
	  +---+---+---+---+---+---+---+---+
	6 |   |   |   |   |   |   |   |   |
	  +---+---+---+---+---+---+---+---+
	5 |   |   |   |   |   |   |   |   |
	  +---+---+---+---+---+---+---+---+
	4 |   |   |   |   |   |   |   |   |
	  +---+---+---+---+---+---+---+---+
	3 |   |   |   |   |   |   |   |   |
	  +---+---+---+---+---+---+---+---+
	2 | P | P | P | P | P | P | P | P |
	  +---+---+---+---+---+---+---+---+
	1 | R | N | B | Q | K | B | N | R |
	  +---+---+---+---+---+---+---+---+
	    a   b   c   d   e   f   g   h

