package main

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/adyzng/go-duka/core"
	"github.com/adyzng/go-duka/csv"
	"github.com/adyzng/go-duka/fxt4"
	"github.com/adyzng/go-duka/hst"
)

var (
	tfRegexp = regexp.MustCompile(`(M|H|D|W|MN)(\d+)`)
	tfMinute = map[string]uint32{
		"M":  1,
		"H":  60,
		"D":  24 * 60,
		"W":  7 * 24 * 60,
		"MN": 30 * 24 * 60,
	}
)

// FormatConvert bi5 to csv/hst/fxt
type FormatConvert struct {
	option AppOption
	tfs    []*Timeframe // M1, M5, M15, M30, H1, H4, D1, W1, MN

}

func NewConvert(opt *AppOption) *FormatConvert {
	tfs := make([]*Timeframe, 0)
	for _, period := range strings.Split(opt.Periods, ",") {
		timeframe, _ := ParseTimeframe(strings.Trim(period, " \t\r\n"))

		var out core.Convertor
		switch opt.Format {
		case "csv":
			out = csv.New(opt.Start, opt.Symbol, opt.Folder, opt.CsvHeader)
			break
		case "fxt":
			out = fxt4.NewFxtFile(timeframe, opt.Spread, opt.Mode, opt.Folder, opt.Symbol)
			break
		case "hst":
			out = hst.NewHST(timeframe, opt.Spread, opt.Symbol, opt.Folder)
			break
		default:
			log.Error("unsupported format %s.", opt.Format)
			return nil
		}

		tfs = append(tfs, NewTimeframe(period, opt.Symbol, out))
	}

	return &FormatConvert{
		option: *opt,
		tfs:    tfs,
	}
}

// Timeframe wrapper of tick data in timeframe like: M1, M5, M15, M30, H1, H4, D1, W1, MN
//
type Timeframe struct {
	deltaTimestamp uint32 // unit second
	startTimestamp uint32 // unit second
	endTimestamp   uint32 // unit second
	timeframe      uint32 // Period of data aggregation in minutes
	period         string // M1, M5, M15, M30, H1, H4, D1, W1, MN
	symbol         string

	chTicks chan *core.TickData
	close   chan struct{}
	out     core.Convertor
}

// ParseTimeframe from input string
//
func ParseTimeframe(period string) (uint32, string) {
	// M15 => [M15 M 15]
	if ss := tfRegexp.FindStringSubmatch(period); len(ss) == 3 {
		n, _ := strconv.Atoi(ss[2])
		for key, val := range tfMinute {
			if key == ss[1] {
				return val * uint32(n), ss[0]
			}
		}
	}
	return 1, "M1" // M1 by default
}

// NewTimeframe create an new timeframe
func NewTimeframe(period, symbol string, out core.Convertor) *Timeframe {
	min, str := ParseTimeframe(period)
	tf := &Timeframe{
		deltaTimestamp: min * 60,
		timeframe:      min,
		period:         str,
		symbol:         symbol,
		out:            out,
		chTicks:        make(chan *core.TickData, 1024),
		close:          make(chan struct{}, 1),
	}

	go tf.worker()
	return tf
}

// PackTicks receive original tick data
func (tf *Timeframe) PackTicks(ticks []*core.TickData) error {
	for _, tick := range ticks {
		select {
		case tf.chTicks <- tick:
			break
		}
	}
	return nil
}

// Finish wait convert finish
func (tf *Timeframe) Finish() {
	<-tf.close
}

// worker thread
func (tf *Timeframe) worker() error {
	maxCap := 1024
	startTime := time.Now()
	barTicks := make([]*core.TickData, 0, maxCap)

	for tick := range tf.chTicks {
		// Beginning of the bar's timeline.
		tickSeconds := uint32(tick.Timestamp / 1000)
		tickBarTime := tickSeconds - tickSeconds%tf.deltaTimestamp

		//Determines the end of the current bar.
		if tickSeconds >= tf.endTimestamp {
			// output one bar data
			if len(barTicks) > 0 {
				tf.out.PackTicks(tickBarTime, barTicks[:])
				barTicks = barTicks[:0]
			}

			// Next bar's timeline will begin from this new tick's bar
			tf.startTimestamp = tickBarTime
			tf.endTimestamp = tf.startTimestamp + tf.deltaTimestamp

			// start next round bar
			barTicks = append(barTicks, tick)

		} else {
			// Tick is within the current bar's timeline, queue it
			barTicks = append(barTicks, tick)
		}
	}

	log.Info("%s : %s convert completed. Time cost: %v.", tf.symbol, tf.period, time.Since(startTime))
	close(tf.close)
	return nil
}
