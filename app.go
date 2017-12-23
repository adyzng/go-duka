package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/adyzng/go-duka/bi5"
	"github.com/adyzng/go-duka/core"
	"github.com/adyzng/go-duka/csv"
	"github.com/adyzng/go-duka/misc"
)

var (
	log = misc.NewLogger("App", 2)
)

// DukaApp used to download source tick data
//
type DukaApp struct {
	Option AppOption
	Output core.Convertor
}

// AppOption download options
//
type AppOption struct {
	Start     time.Time
	End       time.Time
	Symbol    string
	Format    string
	Folder    string
	Periods   string
	Spread    uint32
	Mode      uint32
	Timeframe uint32
	Convert   bool
	CsvDump   bool
	CsvHeader bool
}

type hReader struct {
	Bi5  *bi5.Bi5
	DayH time.Time
	Data []byte
}

// ParseOption parse input command line
//
func ParseOption(args argsList) (*AppOption, error) {
	var (
		err error
		opt AppOption
	)
	if args.Symbol == "" {
		err = fmt.Errorf("Invalid symbol parameter")
		return nil, err
	}
	if opt.Start, err = time.ParseInLocation("2006-01-02", args.Start, time.UTC); err != nil {
		err = fmt.Errorf("invalid start parameter")
		return nil, err
	}
	if opt.End, err = time.ParseInLocation("2006-01-02", args.End, time.UTC); err != nil {
		err = fmt.Errorf("invalid end parameter")
		return nil, nil
	}
	if opt.End.Unix() <= opt.Start.Unix() {
		err = fmt.Errorf("invalid end parameter which shouldn't early then start")
		return nil, err
	}

	if opt.Folder, err = filepath.Abs(args.Folder); err != nil {
		err = fmt.Errorf("invalid destination folder")
		return nil, err
	}
	if err = os.MkdirAll(opt.Folder, 666); err != nil {
		err = fmt.Errorf("create destination folder failed: %v", err)
		return nil, err
	}

	if args.Period != "" {
		args.Period = strings.ToUpper(args.Period)
		if !tfRegexp.MatchString(args.Period) {
			err = fmt.Errorf("invalid timeframe value: %s", args.Period)
			return nil, err
		}
		opt.Periods = args.Period
	}

	opt.Symbol = strings.ToUpper(args.Symbol)
	opt.CsvHeader = args.Header
	opt.CsvDump = true
	opt.Format = args.Format

	return &opt, nil
}

// Execute download source bi5 tick data from dukascopy
//
func (duka *DukaApp) Execute() error {
	var err error
	startTime := time.Now()

	//
	// 按天下载，每天24小时的数据由24个goroutine并行下载
	//
	for day := duka.Option.Start; day.Unix() < duka.Option.End.Unix(); day = day.Add(24 * time.Hour) {
		//
		//  周六没数据，跳过
		//
		if day.Weekday() == time.Saturday {
			log.Warn("Skip Saturday %s.", day.Format("2006-01-02"))
			continue
		}
		//
		// 下载，解析，存储
		//
		if err = duka.saveRaw(day, duka.fetchDay(day)); err != nil {
			break
		}
		log.Info("Finished %s %s.", duka.Option.Symbol, day.Format("2006-01-02"))
	}

	log.Info("Time cost: %v.", time.Since(startTime))
	return err
}

func (duka *DukaApp) fetchDay(day time.Time) <-chan *hReader {
	ch := make(chan *hReader, 24)
	opt := &duka.Option

	go func() {
		defer close(ch)
		var wg sync.WaitGroup

		for hour := 0; hour < 24; hour++ {
			wg.Add(1)
			go func(h int) {
				defer wg.Done()
				dayH := day.Add(time.Duration(h) * time.Hour)
				bi5File := bi5.New(dayH, opt.Symbol, opt.Folder)

				var (
					str  string
					err  error
					data []byte
				)
				if opt.Convert {
					str = "Load Bi5"
					data, err = bi5File.Load()
				} else {
					str = "Download Bi5"
					data, err = bi5File.Download()
				}

				if err != nil {
					log.Error("%s, %s failed: %v.", str, dayH.Format("2006-01-02:15H"))
					return
				}
				if len(data) > 0 {
					select {
					case ch <- &hReader{Data: data[:], DayH: dayH, Bi5: bi5File}:
						break
					}
				}
			}(hour)
		}

		wg.Wait()
		log.Trace("%s:%s download complete.", duka.Option.Symbol, day.Format("2006-01-02"))
	}()

	return ch
}

func (duka *DukaApp) saveRaw(day time.Time, chData <-chan *hReader) error {
	var (
		err     error
		dest    string
		csvFile *csv.CsvDump
		opt     = &duka.Option
	)

	for data := range chData {
		if dest == "" {
			y, m, d := day.Date()
			subDir := fmt.Sprintf("%s/%04d/%02d/%02d", opt.Symbol, y, m, d)

			dest = filepath.Join(opt.Folder, subDir)
			if err = os.MkdirAll(dest, 666); err != nil {
				log.Error("Create folder (%s) failed: %v.", dest, err)
				return err
			}

			if opt.CsvDump {
				csvFile = csv.New(day, opt.Symbol, dest, opt.CsvHeader)
				defer csvFile.Finish()
			}
		}

		// save bi5 by hour
		bi5File := data.Bi5

		// decode bi5
		if ticks, err := bi5File.Decode(data.Data[:]); err != nil {
			log.Error("Decode bi5 %s: %s failed: %v.", opt.Symbol, data.DayH.Format("2006-01-02:15H"), err)
			continue
		} else {
			if opt.CsvDump && len(ticks) > 0 {
				csvFile.PackTicks(0, ticks[:])
			}
		}

		if !opt.Convert {
			// save bi5 source data
			if err := bi5File.Save(data.Data[:]); err != nil {
				log.Error("Save Bi5 %s: %s failed: %v.", opt.Symbol, data.DayH.Format("2006-01-02:15H"), err)
				continue
			}
		}
	}

	if err != nil {
		log.Warn("%s:%s partial complete.", opt.Symbol, day.Format("2006-01-02"))
	} else {
		log.Trace("%s:%s complete.", opt.Symbol, day.Format("2006-01-02"))
	}
	return err
}
