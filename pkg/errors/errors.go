package errors

import (
	"errors"
)

var (
	ErrQuoteNotFound = errors.New("no quote was found")
)
