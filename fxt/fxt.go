package fxt

import (
	"github.com/adyzng/go-duka/core"
)

type fxtTick struct {
	BarTimestamp  int32
	padding       int32
	Open          float64
	High          float64
	Low           float64
	Close         float64
	Volume        uint64
	TickTimestamp int32
	LaunchExpert  int32
}

func convertToTxtTick(tick *core.TickData) *fxtTick {
	return &fxtTick{
		BarTimestamp:  int32(tick.Time.Unix()),
		TickTimestamp: int32(tick.Time.Unix()),
		Open:          tick.Bid,
		High:          tick.Bid,
		Low:           tick.Bid,
		Close:         tick.Bid,
		Volume:        tick.VolumeBid,
	}
}
