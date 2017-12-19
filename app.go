package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adyzng/duka/bi5"
	"github.com/adyzng/duka/csv"
	"github.com/adyzng/duka/download"
	"github.com/adyzng/duka/misc"
)

var (
	log = misc.NewLogger("App", 2)
)

type DukaOption struct {
	Start       time.Time
	End         time.Time
	Symbol      string
	Format      string
	Destination string
}

func App(opt DukaOption) {
	type hReader struct {
		Hour int
		Day  time.Time
		Data []byte
	}

	// dukascopy downloader
	duka := download.NewDukaDownloader()
	startTime := time.Now()

	for day := opt.Start; day.Unix() < opt.End.Unix(); day = day.Add(24 * time.Hour) {
		y, m, d := day.Date()
		chClose := make(chan struct{}, 1)
		chReaders := make(chan *hReader, 24)

		if day.Weekday() == time.Saturday {
			log.Trace("Skip Saturday %s.", day.Format("2006-01-02"))
			continue
		}

		go func() {
			var wg sync.WaitGroup
			for h := 0; h < 24; h++ {
				wg.Add(1)
				go func(hour int) {
					defer wg.Done()
					URL := fmt.Sprintf(download.DukaTmplURL, opt.Symbol, y, m, d, hour)
					if data, err := duka.Download(URL); err == nil {
						chReaders <- &hReader{
							Data: data,
							Hour: hour,
							Day:  day.Add(time.Duration(hour) * time.Hour),
						}
					} else {
						log.Error("Duka download %s failed.", URL)
					}
				}(h)
			}

			wg.Wait()
			close(chReaders)
			log.Info("%s:%s download complete.", opt.Symbol, day.Format("2006-01-02"))
		}()

		go func() {
			defer close(chClose)

			subDir := fmt.Sprintf("%s/%04d/%02d/%02d", opt.Symbol, y, m, d)
			dest := filepath.Join(opt.Destination, subDir)

			if err := os.MkdirAll(dest, 666); err != nil {
				log.Error("Create folder (%s/%s) failed: %v.", opt.Destination, dest, err)
				return
			}

			// save csv by day
			csvFile := csv.New(day, opt.Symbol, dest)

			for chr := range chReaders {
				// save bi5 by hour
				bi5File := bi5.New(chr.Day, chr.Hour, opt.Symbol, dest)

				if ticks, err := bi5File.Decode(bytes.NewBuffer(chr.Data[:])); err != nil {
					log.Error("Decode bi5 %s:%s failed: %v.", opt.Symbol, chr.Day.Format("2006-01-02:15H"))
					continue
				} else {
					csvFile.AddTicks(ticks)
				}

				if err := bi5File.Save(bytes.NewBuffer(chr.Data[:])); err != nil {
					log.Error("Save Bi5 %s:%s failed: %v.", opt.Symbol, chr.Day.Format("2006-01-02:15H"))
					continue
				}
			}

			if err := csvFile.Save(nil); err != nil {
				log.Error("Save CSV %s:%s failed: %v.", opt.Symbol, day.Format("2006-01-02"), err)
			}
		}()

		<-chClose
		log.Info("%s:%s decode complete.", opt.Symbol, day.Format("2006-01-02"))
	}

	log.Info("Completed. Time Cost: %v.", time.Since(startTime))
}
