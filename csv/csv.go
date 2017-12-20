package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/adyzng/go-duka/core"
	"github.com/adyzng/go-duka/misc"
)

var (
	ext = "csv"
	log = misc.NewLogger("CSV", 3)
)

// CsvDump save csv format
type CsvDump struct {
	day    time.Time
	dest   string
	symbol string
	ticks  []*core.TickData
}

func New(day time.Time, symbol, dest string) *CsvDump {
	return &CsvDump{
		day:    day,
		dest:   dest,
		symbol: symbol,
	}
}

func (c *CsvDump) Save(r io.Reader) error {
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

	// sort by time
	sort.Slice(c.ticks, func(i, j int) bool {
		return c.ticks[i].Time.Before(c.ticks[j].Time)
	})

	for _, tick := range c.ticks {
		if err := csv.Write(tick.ToString()); err != nil {
			log.Error("Write CSV %s failed: %v.", fpath, err)
			return err
		}
	}

	log.Trace("Saved file %s with %d ticks.", fpath, len(c.ticks))
	return nil
}

func (c *CsvDump) AddTicks(ticks []*core.TickData) {
	if len(ticks) > 0 {
		c.ticks = append(c.ticks, ticks...)
	}
}
