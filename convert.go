package main

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/adyzng/go-duka/core"
	"github.com/adyzng/go-duka/csv"
	"github.com/adyzng/go-duka/fxt4"
	"github.com/adyzng/go-duka/hst"
	"github.com/go-clog/clog"
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

func NewOutputs(opt *AppOption) []core.Converter {
	outs := make([]core.Converter, 0)
	for _, period := range strings.Split(opt.Periods, ",") {
		var format core.Converter
		timeframe, _ := ParseTimeframe(strings.Trim(period, " \t\r\n"))

		switch opt.Format {
		case "csv":
			format = csv.New(opt.Start, opt.End, opt.CsvHeader, opt.Symbol, opt.Folder)
			break
		case "fxt":
			format = fxt4.NewFxtFile(timeframe, opt.Spread, opt.Mode, opt.Folder, opt.Symbol)
			break
		case "hst":
			format = hst.NewHST(timeframe, opt.Spread, opt.Symbol, opt.Folder)
			break
		default:
			log.Error("unsupported format %s.", opt.Format)
			return nil
		}

		outs = append(outs, NewTimeframe(period, opt.Symbol, format))
	}
	return outs
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
	out     core.Converter
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
func NewTimeframe(period, symbol string, out core.Converter) core.Converter {
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
func (tf *Timeframe) PackTicks(barTimestamp uint32, ticks []*core.TickData) error {
	for _, tick := range ticks {
		select {
		case tf.chTicks <- tick:
			break
		}
	}
	return nil
}

// Finish wait convert finish
func (tf *Timeframe) Finish() error {
	<-tf.close
	return tf.out.Finish()
}

// worker thread
func (tf *Timeframe) worker() error {
	maxCap := 1024
	barTicks := make([]*core.TickData, 0, maxCap)

	defer func() {
		clog.Info("%s %s convert completed.", tf.symbol, tf.period)
		close(tf.close)
	}()

	var tickSeconds uint32
	var tickBarTime uint32

	for tick := range tf.chTicks {
		// Beginning of the bar's timeline.
		tickSeconds = uint32(tick.Timestamp / 1000)
		tickBarTime = tickSeconds - tickSeconds%tf.deltaTimestamp

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

	if len(barTicks) > 0 {
		tf.out.PackTicks(tickBarTime, barTicks[:])
	}

	return nil
}
