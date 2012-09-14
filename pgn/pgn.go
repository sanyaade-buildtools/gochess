package pgn

import "../chess"

import (
	"bufio"
	"os"
	"bytes"
	"regexp"
)

type PGN struct {
	Games []Game                // all the games in the PGN
}

type Game struct {
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

func Parse(filename string) (*PGN, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	// create a line reader and the pgn
	reader := bufio.NewReader(file)
	pgn := new(PGN)

	for {
		if err = pgn.ParseGame(); err != nil {
			break
		}
	}

	return pgn, err
}

func (pgn *PGN) ParseGame(reader *Reader) error {
	var err error

	if err = pgn.ParseTagPairs(); err != nil { return err }

	return nil
}

func (pgn *PGN) ParseTagPairs(reader *Reader) error {
	pgn.Header = make(map[string]string)

	// header match expression
	re := regexp.Compile("\\[([^\\s\t]+)\\s*\"([^\"]*)\"\\]")

	for {
		line, prefix, err := reader.ReadLine()

		if err != nil {
			return nil, err
		}

		match := re.FindSubmatch(line)

		if match == nil {
			return header, nil
		}
	}

	return header, nil
}
