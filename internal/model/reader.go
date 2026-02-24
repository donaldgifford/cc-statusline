package model

import (
	"encoding/json"
	"errors"
	"io"
)

// ErrMalformedJSON indicates the stdin input was not valid JSON.
var ErrMalformedJSON = errors.New("malformed JSON input")

// ReadStatus reads and parses the Claude Code stdin JSON payload.
// Returns a zero-value StatusData (not an error) when stdin is empty.
// Returns ErrMalformedJSON when the input is not valid JSON.
func ReadStatus(r io.Reader) (*StatusData, error) {
	var data StatusData
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		if errors.Is(err, io.EOF) {
			return &StatusData{}, nil
		}
		return nil, ErrMalformedJSON
	}
	return &data, nil
}
