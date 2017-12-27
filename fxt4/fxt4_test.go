package fxt4

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/adyzng/go-duka/core"
)

func TestFxtFile(t *testing.T) {
	fxt := NewFxtFile(1, 20, 0, "D:\\Data", "EURUSD")
	fxt.PackTicks(0, []*core.TickData{&core.TickData{}})
}

func TestHeader(t *testing.T) {
	//fname := `F:\tester-ok\EURUSD1_0.fxt`
	fname := `E:\test\EURUSD5_0.fxt`

	fh, err := os.OpenFile(fname, os.O_RDONLY, 666)
	if err != nil {
		t.Fatalf("Open fxt file failed: %v.\n", err)
	}
	defer fh.Close()

	bs := make([]byte, headerSize)
	n, err := fh.Read(bs[:])
	if err != nil || n != headerSize {
		t.Fatalf("Read fxt header failed: %v.\n", err)
	}

	var h FXTHeader
	err = binary.Read(bytes.NewBuffer(bs[:]), binary.LittleEndian, &h)
	if err != nil {
		t.Fatalf("Decode fxt header failed: %v.\n", err)
	}

	tickBs := make([]byte, tickSize)
	for {
		n, err = fh.Read(tickBs[:tickSize])
		if err == io.EOF {
			break
		}

		if n != tickSize || err != nil {
			t.Errorf("Read tick data failed: %v.\n", err)
			break
		}

		var tick FxtTick
		err = binary.Read(bytes.NewBuffer(tickBs[:]), binary.LittleEndian, &tick)
		if err != nil {
			t.Errorf("Decode tick data failed: %v.\n", err)
			break
		}

		fmt.Println(tick)
	}
	fmt.Printf("Header:\n%+v\n", h)
}
