package fxt4

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/adyzng/go-duka/core"
	"github.com/adyzng/go-duka/misc"
)

var (
	log = misc.NewLogger("FXT", 3)
)

// FxtFile define fxt file format
//
// Refer: https://github.com/EA31337/MT-Formats
// FXT file should be placed in the tester/history directory. name format is SSSSSSPP_M.fxt where:
//		SSSSSS - symbol name same as in symbol field in the header
//		PP - timeframe period must be correspond with period field in the header
//		M - model number (0,1 or 2)
//
type FxtFile struct {
	fpath          string
	symbol         string
	model          uint32
	header         *FXTHeader
	firstUniBar    *FxtTick
	lastUniBar     *FxtTick
	deltaTimestamp uint32
	endTimestamp   uint32
	timeframe      uint32
	barCount       int32
	tickCount      int64
	chTicks        chan *FxtTick
	chClose        chan struct{}
}

// NewFxtFile create an new fxt file instance
func NewFxtFile(timeframe, spread, model uint32, dest, symbol string) *FxtFile {
	fn := fmt.Sprintf("%s%d_%d.fxt", symbol, timeframe, model)
	fxt := &FxtFile{
		header:         NewHeader(405, symbol, timeframe, spread, model),
		fpath:          filepath.Join(dest, fn),
		chTicks:        make(chan *FxtTick, 1024),
		chClose:        make(chan struct{}, 1),
		deltaTimestamp: timeframe * 60,
		timeframe:      timeframe,
		symbol:         symbol,
		model:          model,
	}

	go fxt.worker()
	return fxt
}

func (f *FxtFile) worker() error {
	defer func() {
		close(f.chClose)
		log.Info("M%d Saved Bar: %d, Ticks: %d.", f.timeframe, f.barCount, f.tickCount)
	}()

	fxt, err := os.OpenFile(f.fpath, os.O_CREATE|os.O_TRUNC, 666)
	if err != nil {
		log.Fatal("Create file %s failed: %v.", f.fpath, err)
		return err
	}

	defer fxt.Close()
	bu := bytes.NewBuffer(make([]byte, 0, headerSize))

	//
	// convert FXT header
	//
	if err = binary.Write(bu, binary.LittleEndian, f.header); err != nil {
		log.Error("Convert FXT header failed: %v.", err)
		return err
	}
	// write FXT file
	if _, err := fxt.Write(bu.Bytes()); err != nil {
		log.Error("Write FXT header failed: %v.", err)
		return err
	}

	for tick := range f.chTicks {

		if tick.BarTimestamp > tick.TickTimestamp {
			log.Fatal("Tick(%v)", tick)
		}

		bu.Reset()
		//
		//  write tick data
		//
		if err = binary.Write(bu, binary.LittleEndian, tick); err != nil {
			log.Error("Pack tick failed: %v.", err)
			break
		}
		if _, err = fxt.Write(bu.Bytes()); err != nil {
			log.Error("Write fxt tick (%x) failed: %v.", bu.Bytes(), err)
			break
		}

		if f.firstUniBar == nil {
			f.firstUniBar = tick
		}
		f.lastUniBar = tick
	}
	return err
}

func (f *FxtFile) PackTicks(barTimestemp uint32, ticks []*core.TickData) error {

	if len(ticks) == 0 {
		return nil
	}

	var (
		op = ticks[0].Bid
		hi = ticks[0].Bid
		lo = ticks[0].Bid
		vo = math.Max(ticks[0].VolumeBid, 1)
	)

	for _, tick := range ticks {
		ft := &FxtTick{
			BarTimestamp:  uint32(barTimestemp),
			TickTimestamp: uint32(tick.Timestamp / 1000),
			Open:          op,
			High:          math.Max(tick.Bid, hi),
			Low:           math.Min(tick.Bid, lo),
			Close:         tick.Bid,
			Volume:        uint64(vo),
			//RealSpread:    uint32(tick.Ask - tick.Bid/f.header.PointSize),
			//LaunchExpert:  3,
		}
		vo = vo + tick.VolumeBid
		f.chTicks <- ft
		f.tickCount++
	}
	if f.endTimestamp != barTimestemp {
		f.barCount++
		f.endTimestamp = barTimestemp
	}
	return nil
}

func (f *FxtFile) adjustHeader() error {
	if f.barCount == 0 {
		return nil
	}

	fxt, err := os.OpenFile(f.fpath, os.O_RDWR, 666)
	if err != nil {
		log.Fatal("Open file %s failed: %v.", f.fpath, err)
		return err
	}
	defer fxt.Close()

	// first part
	if _, err := fxt.Seek(216, os.SEEK_SET); err == nil {
		d := struct {
			BarCount          int32  // Total bar count
			BarStartTimestamp uint32 // Modelling start date - date of the first tick.
			BarEndTimestamp   uint32 // Modelling end date - date of the last tick.
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
			BarStartTimestamp uint32 // Tester start date - date of the first tick.
			BarEndTimestamp   uint32 // Tester end date - date of the last tick.
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
	<-f.chClose
	return f.adjustHeader()
}
