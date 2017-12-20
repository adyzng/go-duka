package fxt4

import (
	"testing"

	"github.com/adyzng/go-duka/core"
)

func TestFxtFile(t *testing.T) {
	fxt := NewFxtFile(1, 20, 0, "D:\\Data", "EURUSD")
	fxt.AddTicks([]*core.TickData{&core.TickData{}})
}
