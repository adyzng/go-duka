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
	barCount       int32
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
	if err = binary.Write(bu, binary.LittleEndian, f.header); err != nil {
		log.Error("Write FXT header failed: %v.", err)
		return err
	}

	for tick := range f.chTicks {
		bu.Reset()
		//
		//  write tick data
		//
		if err = binary.Write(bu, binary.LittleEndian, tick); err != nil {
			log.Error("Pack tick failed: %v.", err)
			break
		}

		if f.firstUniBar == nil {
			f.firstUniBar = tick
		}
		f.lastUniBar = tick

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
		f.chTicks <- &fxtTick{
			BarTimestamp:  int32(barTimestemp),
			TickTimestamp: int32(tick.Timestamp / 1000),
			Open:          tick.Bid,
			High:          tick.Bid,
			Low:           tick.Bid,
			Close:         tick.Bid,
			Volume:        uint64(tick.VolumeBid),
		}
		//f.chTicks <- convertToFxtTick(tick)
	}
	f.barCount++
	return nil
}

func (f *FxtFile) adjustHeader() error {
	fxt, err := os.OpenFile(f.fpath, os.O_RDWR, 666)
	if err != nil {
		log.Fatal("Open file %s failed: %v.", f.fpath, err)
		return err
	}
	defer fxt.Close()

	// first part
	if _, err := fxt.Seek(216, os.SEEK_SET); err == nil {
		d := struct {
			BarCount          int32 // Total bar count
			BarStartTimestamp int32 // Modelling start date - date of the first tick.
			BarEndTimestamp   int32 // Modelling end date - date of the last tick.
		}{
			f.barCount,
			f.firstUniBar.BarTimestamp,
			f.lastUniBar.BarTimestamp,
		}

		bu := new(bytes.Buffer)
		if err = binary.Write(bu, binary.LittleEndian, &d); err == nil {
			_, err = fxt.Write(bu.Bytes())
		}
		if err != nil {
			log.Error("Adjust FXT header 1 failed: %v.", err)
			return err
		}
	} else {
		log.Error("File seek 1 failed: %v.", err)
		return err
	}

	// end part
	if _, err := fxt.Seek(472, os.SEEK_SET); err == nil {
		d := struct {
			BarStartTimestamp int32 // Tester start date - date of the first tick.
			BarEndTimestamp   int32 // Tester end date - date of the last tick.
		}{
			f.firstUniBar.BarTimestamp,
			f.lastUniBar.BarTimestamp,
		}
		bu := new(bytes.Buffer)
		if err = binary.Write(bu, binary.LittleEndian, &d); err == nil {
			_, err = fxt.Write(bu.Bytes())
		}
		if err != nil {
			log.Error("Adjust FXT header 2 failed: %v.", err)
			return err
		}
	} else {
		log.Error("File seek 2 failed: %v.", err)
		return err
	}

	return nil
}

func (f *FxtFile) Finish() error {
	close(f.chTicks)
	f.wg.Wait()
	return f.adjustHeader()
}
