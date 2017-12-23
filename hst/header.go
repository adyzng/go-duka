package hst

import (
	"time"

	"github.com/adyzng/go-duka/misc"
)

const (
	v401        = uint32(401)
	barBytes    = 60
	headerBytes = 148
)

// Header structure for hst version 401 (148 bytes).
//
type Header struct {
	Version   uint32     //   0    4   HST version (401)
	Copyright [64]byte   //   4   64   Copyright info
	Symbol    [12]byte   //  68   12   Forex symbol
	Period    uint32     //  80    4   Symbol timeframe
	Digits    uint32     //  84    4   The amount of digits after decimal point in the symbol
	TimeSign  uint32     //  88    4   Time of sign (database creation)
	LastSync  uint32     //  92    4   Time of last synchronization
	unused    [13]uint32 //  96   52   unused
}

// BarData wrap the bar data inside hst (60 Bytes)
//
type BarData struct {
	CTM        uint32  //   0   4   current time in seconds
	_          uint32  //   4   4   for padding only
	Open       float64 //   8   8   OLHCV
	Low        float64 //  16   8   L
	High       float64 //  24   8   H
	Close      float64 //  32   8   C
	Volume     uint64  //  40   8   V
	Spread     uint32  //  48   4
	RealVolume uint64  //  52   8
}

// NewHeader for hst version 401
//
func NewHeader(timeframe uint32, symbol string) *Header {
	h := &Header{
		TimeSign: uint32(time.Now().UTC().Unix()),
		Version:  v401,
		Period:   timeframe,
		Digits:   5, // Digits, using the default value of HST format
	}

	misc.ToFixBytes(h.Symbol[:], symbol)
	misc.ToFixBytes(h.Copyright[:], "(C)opyright 2017, MetaQuotes Software Corp.")
	return h
}

// ToBytes convert header to fix bytes array
//
func (h *Header) ToBytes() ([]byte, error) {
	bs, err := misc.PackLittleEndian(headerBytes, h)
	if err != nil {
		log.Error("Failed to convert HST header to bytes array. Error %v.", err)
		return make([]byte, 0), err
	}
	return bs, err
}

// ToBytes convert bar data to fix bytes array
//
func (b *BarData) ToBytes() ([]byte, error) {
	bs, err := misc.PackLittleEndian(barBytes, b)
	if err != nil {
		log.Error("Failed to convert HST Bar data to bytes array. Error %v.", err)
		return make([]byte, 0), err
	}
	return bs, err
}
