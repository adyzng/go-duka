package main

import (
	"flag"
	"fmt"
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

type argsList struct {
	Verbose bool
	Header  bool
	Symbol  string
	Folder  string
	Format  string
	Period  string
	Start   string
	End     string
}

func main() {
	args := argsList{}
	start := time.Now().Format("2006-01-02")
	end := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	flag.StringVar(&args.Period, "timeframe", "M1", "timeframe values: M1, M5, M15, M30, H1, H4, D1, W1, MN")
	flag.StringVar(&args.Symbol, "symbol", "", "symbol list using format, like: EURUSD EURGBP")
	flag.StringVar(&args.Start, "start", start, "start date format YYYY-MM-DD (default today)")
	flag.StringVar(&args.End, "end", end, "end date format YYYY-MM-DD (default today)")
	flag.StringVar(&args.Folder, "folder", ".", "destination folder (default .)")
	flag.BoolVar(&args.Verbose, "verbose", false, "verbose output trace log")
	flag.BoolVar(&args.Header, "header", false, "save csv with header")
	flag.Parse()

	if args.Verbose {
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

	opt, err := ParseOption(args)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("    Folder: %s\n", opt.Folder)
	fmt.Printf("    Symbol: %s\n", opt.Symbol)
	fmt.Printf(" Timeframe: %d\n", opt.Timeframe)
	fmt.Printf(" StartDate: %s\n", opt.Start.Format("2006-01-02"))
	fmt.Printf("   EndDate: %s\n", opt.End.Format("2006-01-02"))
	fmt.Printf("    Format: %s\n", opt.Format)
	fmt.Printf("   DumpCsv: %t\n", opt.CsvDump)
	fmt.Printf(" CsvHeader: %t\n", opt.CsvHeader)

	defer clog.Shutdown()
	app := DukaApp{Option: *opt}
	app.Execute()
}
