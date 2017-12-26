package csv

import (
	"testing"
	"time"
)

func TestCloseChan(t *testing.T) {
	chClose := make(chan struct{}, 1)
	go func() {
		defer close(chClose)
		time.Sleep(2 * time.Second)
		t.Logf("Close chan.\n")
	}()

	<-chClose
	t.Logf("Receive close channel.\n")
}
