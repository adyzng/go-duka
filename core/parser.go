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

// Converter convert raw tick data into different file format
// such as fxt, hst, csv
type Converter interface {
	PackTicks(barTimestamp uint32, ticks []*TickData) error
	Finish() error
}
