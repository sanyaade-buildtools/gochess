package pgn

import "../chess"

import (
	"regexp"
)

type errno int

const (
	InvalidMoveString = errno(1)
)

var errmap = map[errno]string{
	InvalidMoveString: "Invalid move string",
}

// regular expression for parsing a move string
