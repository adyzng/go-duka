package fxt4

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adyzng/go-duka/core"
	"github.com/adyzng/go-duka/misc"
)

var (
	log = misc.NewLogger("FXT", 3)
)

type fxtTick struct {
	BarTimestamp  int32
	padding       int32
	Open          float64
	High          float64
	Low           float64
	Close         float64
	Volume        uint64
	TickTimestamp int32
	LaunchExpert  int32
}

func convertToFxtTick(tick *core.TickData) *fxtTick {
	return &fxtTick{
		BarTimestamp:  int32(tick.Time.Unix()),
		TickTimestamp: int32(tick.Time.Unix()),
		Open:          tick.Bid,
		High:          tick.Bid,
		Low:           tick.Bid,
		Close:         tick.Bid,
		Volume:        tick.VolumeBid,
	}
}

// FxtFile define fxt file format
//
// Refer: https://github.com/EA31337/MT-Formats
// FXT file should be placed in the tester/history directory. name format is SSSSSSPP_M.fxt where:
//		SSSSSS - symbol name same as in symbol field in the header
//		PP - timeframe period must be correspond with period field in the header
//		M - model number (0,1 or 2)
//
type FxtFile struct {
	firstUniBar *fxtTick
	lastUniBar  *fxtTick
	chTicks     chan *fxtTick
	fpath       string

	deltaTimestamp uint32
	endTimestamp   uint32
	barCount       uint64
}

// NewFxtFile create an new fxt file instance
func NewFxtFile(timeframe, spread, model uint32, dest, symbol string) *FxtFile {
	fn := fmt.Sprintf("%s%02d_%d.fxt", symbol, timeframe, model)
	fxt, err := os.OpenFile(filepath.Join(dest, fn), os.O_CREATE|os.O_TRUNC, 666)
	if err != nil {
		log.Fatal("Create file %s failed: %v.", fn, err)
		return nil
	}
	defer fxt.Close()

	return &FxtFile{
		fpath:          filepath.Join(dest, fn),
		chTicks:        make(chan *fxtTick, 1024),
		deltaTimestamp: timeframe * 60,
	}
}

func (f *FxtFile) worker() error {
	fxt, err := os.OpenFile(f.fpath, os.O_CREATE|os.O_TRUNC, 666)
	if err != nil {
		log.Fatal("Create file %s failed: %v.", f.fpath, err)
		return err
	}

	defer fxt.Close()
	bu := bytes.NewBuffer(make([]byte, 4096))

	for tick := range f.chTicks {
		bu.Reset()
		if err = binary.Write(bu, binary.BigEndian, tick); err != nil {
			log.Error("Pack tick failed: %v.", err)
			break
		}

		f.lastUniBar = tick
		if f.firstUniBar == nil {
			f.firstUniBar = tick
		}

		if _, err = fxt.Write(bu.Bytes()); err != nil {
			log.Error("Write fxt tick (%x) failed: %v.", bu.Bytes(), err)
			break
		}
	}
}

func (f *FxtFile) AddTicks(ticks []*core.TickData) {
	for _, tick := range ticks {
		f.chTicks <- convertToFxtTick(tick)
	}
}
