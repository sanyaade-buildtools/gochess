package pgn

import "../chess"

import (
	"io/ioutil"
	"regexp"
)

type PGN struct {
	Tags map[string]string      // settings at the top of the file
	Moves [][2]*Move            // list of moves made
	Result int                  // end game result
	Comment string              // optional comment
}

type Move struct {
	Move *chess.Move            // actual move
	Alternative []*chess.Move   // alternative line of moves
	Comment string              // optional comment
}

const (
	InProgress = iota
	Draw
	WhiteWins
	BlackWins
)

var reTagPair, _ = regexp.Compile("^\\[([^\\s\t]+)\\s*\"([^\"]*)\"\\]")
var reWhitespace, _ = regexp.Compile("[\\s\\n]*")

func Parse(filename string) ([]*PGN, error) {
	bytes, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	// create a channel for pgn games
	games := make(chan *PGN)
	pgns := make([]*PGN, 0, 1)

	// start the process of parsing the games
	go ParseGames(games, bytes, &err)

	// slurp in each game as it was read
	for game := range games {
		pgns = append(pgns, game)
	}

	// error is non-nil if a game failed to parse
	return pgns, err
}

func ParseGames(ch chan *PGN, text []byte, err *error) {
	var game *PGN

	for len(text) > 0 {
		if game, *err = ParseGame(&text); *err != nil {
			break
		}

		// write the game
		ch <- game
	}

	close(ch)
}

func ParseGame(text *[]byte) (*PGN, error) {
	var err error

	// create the PGN to parse into
	pgn := new(PGN)

	// parse the various sections
	if err = pgn.ParseTagPairs(text); err != nil { return nil, err }

	return pgn, nil
}

func (pgn *PGN) ParseTagPairs(text *[]byte) error {
	pgn.Tags = make(map[string]string)

	for {
		match := reTagPair.FindSubmatch(*text)

		if match == nil {
			return nil
		}

		// add the tag to the PGN
		pgn.Tags[string(match[1])] = string(match[2])

		// advance the pointer
		*text = (*text)[len(match[0]):]
	}

	return nil
}
