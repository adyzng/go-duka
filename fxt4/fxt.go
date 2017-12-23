package fxt4

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"

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
		BarTimestamp:  int32(tick.Timestamp / 1000),
		TickTimestamp: int32(tick.Timestamp / 1000),
		Open:          tick.Bid,
		High:          tick.Bid,
		Low:           tick.Bid,
		Close:         tick.Bid,
		Volume:        uint64(tick.VolumeBid),
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
	wg             sync.WaitGroup
	header         *FXTHeader
	chTicks        chan *fxtTick
	fpath          string
	firstUniBar    *fxtTick
	lastUniBar     *fxtTick
	deltaTimestamp uint32
	endTimestamp   uint32
	barCount       uint64
}

// NewFxtFile create an new fxt file instance
func NewFxtFile(timeframe, spread, model uint32, dest, symbol string) *FxtFile {
	fn := fmt.Sprintf("%s%02d_%d.fxt", symbol, timeframe, model)
	fxt := &FxtFile{
		header:         NewHeader(405, symbol, timeframe, spread, model),
		fpath:          filepath.Join(dest, fn),
		chTicks:        make(chan *fxtTick, 1024),
		deltaTimestamp: timeframe * 60,
	}

	fxt.wg.Add(1)
	go fxt.worker()
	return fxt
}

func (f *FxtFile) worker() error {
	defer f.wg.Done()

	fxt, err := os.OpenFile(f.fpath, os.O_CREATE|os.O_TRUNC, 666)
	if err != nil {
		log.Fatal("Create file %s failed: %v.", f.fpath, err)
		return err
	}

	defer fxt.Close()
	bu := bytes.NewBuffer(make([]byte, 4096))

	//
	// write FXT header
	//
	if err = binary.Write(bu, binary.BigEndian, f.header); err != nil {
		log.Error("Write FXT header failed: %v.", err)
		return err
	}

	for tick := range f.chTicks {
		bu.Reset()

		//
		//  write tick data
		//
		if err = binary.Write(bu, binary.BigEndian, tick); err != nil {
			log.Error("Pack tick failed: %v.", err)
			break
		}

		f.barCount++
		f.lastUniBar = tick
		if f.firstUniBar == nil {
			f.firstUniBar = tick
		}

		if _, err = fxt.Write(bu.Bytes()); err != nil {
			log.Error("Write fxt tick (%x) failed: %v.", bu.Bytes(), err)
			break
		}
	}

	log.Trace("Total %u ticks write.", f.barCount)
	return err
}

func (f *FxtFile) PackTicks(barTimestemp uint32, ticks []*core.TickData) error {
	for _, tick := range ticks {
		f.chTicks <- convertToFxtTick(tick)
	}
	return nil
}

func (f *FxtFile) Finish() error {
	close(f.chTicks)
	f.wg.Wait()
	return nil
}
