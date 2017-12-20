package core

import (
	"io"
)

// Parser interface used to parse data
type Parser interface {
	Parse(r io.Reader) error
}

// Saver interface used to save data
type Saver interface {
	Save(r io.Reader) error
}
