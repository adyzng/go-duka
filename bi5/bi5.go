package bi5

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/adyzng/go-duka/core"
	"github.com/adyzng/go-duka/misc"
	"github.com/kjk/lzma"
)

var (
	ext = "bi5"
	log = misc.NewLogger("Bi5", 3)
)

type Bi5 struct {
	timeH  time.Time
	dest   string
	symbol string
}

// New create an bi5 saver
func New(day time.Time, hour int, symbol, dest string) *Bi5 {
	return &Bi5{
		timeH:  day.Add(time.Duration(hour) * time.Hour),
		dest:   dest,
		symbol: symbol,
	}
}

func (b *Bi5) Decode(r io.Reader) ([]*core.TickData, error) {
	dec := lzma.NewReader(r)
	defer dec.Close()

	ticksArr := make([]*core.TickData, 0)
	bytesArr := make([]byte, core.TICK_BYTES)

	for {
		n, err := dec.Read(bytesArr[:])
		if err == io.EOF {
			err = nil
			break
		}

		if n != core.TICK_BYTES || err != nil {
			log.Error("LZMA decode failed: %d: %v.", n, err)
			break
		}

		t, err := core.DecodeTickData(bytesArr[:], b.symbol, b.timeH)
		if err != nil {
			log.Error("Decode tick data failed: %v.", err)
			break
		}

		ticksArr = append(ticksArr, t)
	}

	return ticksArr, nil
}

func (b *Bi5) Save(r io.Reader) error {
	subpath := fmt.Sprintf("%02dh.%s", b.timeH.Hour(), ext)
	fpath := filepath.Join(b.dest, subpath)

	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 666)
	if err != nil {
		log.Error("Create file %s failed: %v.", fpath, err)
		return err
	}

	var len int64
	if len, err = io.Copy(f, r); err == nil {
		if len > 0 {
			log.Trace("Saved file %s => %d.", fpath, len)
		}
	} else {
		log.Error("Write file %s failed: %v.", fpath, err)
	}

	f.Close()
	if len == 0 {
		os.Remove(fpath)
	}
	return err
}
