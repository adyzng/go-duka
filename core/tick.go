package core

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"
)

const (
	TICK_BYTES = 20
)

var (
	normSymbols = []string{"USDRUB", "XAGUSD", "XAUUSD"}
)

// TickData ...
// struck.unpack(!IIIff)
// date, ask / point, bid / point, round(volume_ask * 1000000), round(volume_bid * 1000000)
type TickData struct {
	Symbol    string
	Time      time.Time
	Ask       float64
	Bid       float64
	VolumeAsk uint64
	VolumeBid uint64
}

func (t *TickData) ToString() []string {
	return []string{
		t.Time.Format("2006-01-02 15:04:05.000"),
		fmt.Sprintf("%.6f", t.Ask),
		fmt.Sprintf("%.6f", t.Bid),
		fmt.Sprintf("%d", t.VolumeAsk),
		fmt.Sprintf("%d", t.VolumeBid),
	}
}

// DecodeTickData from input data bytes array.
// the valid data array should be at size `TICK_BYTES`.
func DecodeTickData(data []byte, symbol string, timeH time.Time) (*TickData, error) {
	raw := struct {
		TimeMs    int32
		Ask       int32
		Bid       int32
		VolumeAsk float32
		VolumeBid float32
	}{}

	if len(data) != TICK_BYTES {
		return nil, errors.New("invalid length for tick data")
	}

	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.BigEndian, &raw); err != nil {
		return nil, err
	}

	var point float64 = 100000
	for _, sym := range normSymbols {
		if symbol == sym {
			point = 1000
			break
		}
	}

	round := func(f float64) uint64 {
		f += 0.5
		return uint64(math.Floor(f))
	}

	t := TickData{
		Symbol:    symbol,
		Time:      timeH.Add(time.Duration(raw.TimeMs) * time.Millisecond),
		Ask:       float64(raw.Ask) / point,
		Bid:       float64(raw.Bid) / point,
		VolumeAsk: round(float64(raw.VolumeAsk) * 1000000),
		VolumeBid: round(float64(raw.VolumeBid) * 1000000),
	}

	return &t, nil
}
