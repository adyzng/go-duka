package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/adyzng/go-duka/core"
	"github.com/adyzng/go-duka/misc"
)

var (
	ext       = "csv"
	log       = misc.NewLogger("CSV", 3)
	csvHeader = []string{"time", "ask", "bid", "ask_volume", "bid_volume"}
)

// CsvDump save csv format
type CsvDump struct {
	day    time.Time
	dest   string
	symbol string
	header bool
	ticks  []*core.TickData
}

// New Csv file
func New(day time.Time, symbol, dest string, header bool) *CsvDump {
	return &CsvDump{
		day:    day,
		dest:   dest,
		symbol: symbol,
		header: header,
	}
}

// Finish complete csv file writing
//
func (c *CsvDump) Finish() error {
	if len(c.ticks) == 0 {
		return nil
	}

	subpath := fmt.Sprintf("%s-%s.%s", c.symbol, c.day.Format("2006-01-02"), ext)
	fpath := filepath.Join(c.dest, subpath)

	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 666)
	if err != nil {
		log.Error("Create file %s failed: %v.", fpath, err)
		return err
	}
	defer f.Close()

	csv := csv.NewWriter(f)
	defer csv.Flush()

	if c.header {
		csv.Write(csvHeader)
	}

	// sort by time
	sort.Slice(c.ticks, func(i, j int) bool {
		return c.ticks[i].Timestamp < c.ticks[j].Timestamp
	})

	for _, tick := range c.ticks {
		if err := csv.Write(tick.ToString()); err != nil {
			log.Error("Write csv %s failed: %v.", subpath, err)
			return err
		}
	}

	log.Trace("Saved file %s with %d ticks.", subpath, len(c.ticks))
	return nil
}

// PackTicks handle ticks data
//
func (c *CsvDump) PackTicks(barTimestamp uint32, ticks []*core.TickData) error {
	if len(ticks) > 0 {
		c.ticks = append(c.ticks, ticks...)
	}
	return nil
}
