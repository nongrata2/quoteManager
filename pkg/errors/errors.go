package errors

import (
	"errors"
)

var (
	ErrQuoteNotFound = errors.New("no quote was found")
	ErrExecDB        = errors.New("db exec error")
	ErrQuery         = errors.New("db query error")
)
