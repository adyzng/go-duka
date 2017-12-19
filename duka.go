package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-clog/clog"
)

func init() {
	/*
		var fpath string
		if logPath == "" {
			fpath, _ = os.Getwd()
		} else {
			fpath, _ = filepath.Abs(logPath)
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 666); err != nil {
			fmt.Printf("[App] Create log folder failed: %v.", err)
			os.Exit(-1)
		}
		log.Trace("App Path: %s.", fpath)

		log.New(log.FILE, log.FileConfig{
			Level:      log.TRACE,
			Filename:   filepath.Join(fpath, "app.log"),
			BufferSize: 2048,
			FileRotationConfig: log.FileRotationConfig{
				Rotate:  true,
				MaxDays: 30,
				MaxSize: 50 * (1 << 20),
			},
		})
	*/
}

func main() {
	var (
		verbose bool
		symbol  string
		dest    string
		format  string
		start   = time.Now().Format("2006-01-02")
		end     = time.Now().Add(24 * time.Hour).Format("2006-01-02")
	)
	flag.StringVar(&start, "start", start, "start date format YYYY-MM-DD (default today)")
	flag.StringVar(&end, "end", end, "end date format YYYY-MM-DD (default today)")
	flag.StringVar(&dest, "folder", ".", "destination folder (default .)")
	flag.StringVar(&symbol, "symbol", "", "symbol list using format like: EURUSD EURGBP")
	flag.BoolVar(&verbose, "verbose", false, "verbose output trace log")
	flag.Parse()

	var (
		err error
		opt DukaOption
	)
	if opt.Start, err = time.ParseInLocation("2006-01-02", start, time.UTC); err != nil {
		fmt.Println("Invalid start parameter")
		return
	}
	if opt.End, err = time.ParseInLocation("2006-01-02", end, time.UTC); err != nil {
		fmt.Println("Invalid end parameter")
		return
	}
	if opt.End.Unix() <= opt.Start.Unix() {
		fmt.Printf("%s -> %s.\n", opt.Start, opt.End)
		fmt.Println("Invalid end parameter which shouldn't early then start")
		return
	}

	if opt.Destination, err = filepath.Abs(dest); err != nil {
		fmt.Println("Invalid destination folder")
		return
	}
	if err = os.MkdirAll(opt.Destination, 666); err != nil {
		fmt.Printf("Create destination folder failed: %v.\n", err)
		return
	}

	if symbol == "" {
		fmt.Println("Invalid symbol parameter")
		return
	}
	opt.Symbol = strings.ToUpper(symbol)
	opt.Format = format

	if verbose {
		clog.New(clog.CONSOLE, clog.ConsoleConfig{
			Level:      clog.TRACE,
			BufferSize: 100,
		})
	} else {
		clog.New(clog.CONSOLE, clog.ConsoleConfig{
			Level:      clog.INFO,
			BufferSize: 100,
		})
	}

	defer clog.Shutdown()
	App(opt)
}
