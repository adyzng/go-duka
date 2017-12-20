package fxt4

import (
	"strings"
)

/*
struct FXT_HEADER {                                // -- offset ---- size --- description ----------------------------------------------------------------------------
   UINT   version;                                 //         0         4     header version:                             405
   char   description[64];                         //         4        64     copyright/description (szchar)
   char   serverName[128];                         //        68       128     account server name (szchar)
   char   symbol[MAX_SYMBOL_LENGTH+1];             //       196        12     symbol (szchar)
   UINT   period;                                  //       208         4     timeframe in minutes
   UINT   modelType;                               //       212         4     0=EveryTick|1=ControlPoints|2=BarOpen
   UINT   modeledBars;                             //       216         4     number of modeled bars      (w/o prolog)
   UINT   firstBarTime;                            //       220         4     bar open time of first tick (w/o prolog)
   UINT   lastBarTime;                             //       224         4     bar open time of last tick  (w/o prolog)
   BYTE   reserved_1[4];                           //       228         4     (alignment to the next double)
   double modelQuality;                            //       232         8     max. 99.9

   // common parameters                            // ----------------------------------------------------------------------------------------------------------------
   char   baseCurrency[MAX_SYMBOL_LENGTH+1];       //       240        12     base currency (szchar)                     = StringLeft(symbol, 3)
   UINT   spread;                                  //       252         4     spread in points: 0=zero spread            = MarketInfo(MODE_SPREAD)
   UINT   digits;                                  //       256         4     digits                                     = MarketInfo(MODE_DIGITS)
   BYTE   reserved_2[4];                           //       260         4     (alignment to the next double)
   double pointSize;                               //       264         8     resolution, ie. 0.0000'1                   = MarketInfo(MODE_POINT)
   UINT   minLotsize;                              //       272         4     min lot size in centi lots (hundredths)    = MarketInfo(MODE_MINLOT)  * 100
   UINT   maxLotsize;                              //       276         4     max lot size in centi lots (hundredths)    = MarketInfo(MODE_MAXLOT)  * 100
   UINT   lotStepsize;                             //       280         4     lot stepsize in centi lots (hundredths)    = MarketInfo(MODE_LOTSTEP) * 100
   UINT   stopsLevel;                              //       284         4     orders stop distance in points             = MarketInfo(MODE_STOPLEVEL)
   BOOL   pendingsGTC;                             //       288         4     close pending orders at end of day or GTC
   BYTE   reserved_3[4];                           //       292         4     (alignment to the next double)

   // profit calculation parameters                // ----------------------------------------------------------------------------------------------------------------
   double contractSize;                            //       296         8     ie. 100000                                 = MarketInfo(MODE_LOTSIZE)
   double tickValue;                               //       304         8     tick value in quote currency (empty)       = MarketInfo(MODE_TICKVALUE)
   double tickSize;                                //       312         8     tick size (empty)                          = MarketInfo(MODE_TICKSIZE)
   UINT   profitCalculationMode;                   //       320         4     0=Forex|1=CFD|2=Futures                    = MarketInfo(MODE_PROFITCALCMODE)

   // swap calculation parameters                  // ----------------------------------------------------------------------------------------------------------------
   BOOL   swapEnabled;                             //       324         4     if swaps are to be applied
   UINT   swapCalculationMode;                     //       328         4     0=Points|1=BaseCurrency|2=Interest|3=MarginCurrency   = MarketInfo(MODE_SWAPTYPE)
   BYTE   reserved_4[4];                           //       332         4     (alignment to the next double)
   double swapLongValue;                           //       336         8     long overnight swap value                  = MarketInfo(MODE_SWAPLONG)
   double swapShortValue;                          //       344         8     short overnight swap values                = MarketInfo(MODE_SWAPSHORT)
   UINT   tripleRolloverDay;                       //       352         4     weekday of triple swaps                    = WEDNESDAY (3)

   // margin calculation parameters                // ----------------------------------------------------------------------------------------------------------------
   UINT   accountLeverage;                         //       356         4     account leverage                           = AccountLeverage()
   UINT   freeMarginCalculationType;               //       360         4     free margin calculation type               = AccountFreeMarginMode()
   UINT   marginCalculationMode;                   //       364         4     margin calculation mode                    = MarketInfo(MODE_MARGINCALCMODE)
   UINT   marginStopoutLevel;                      //       368         4     margin stopout level                       = AccountStopoutLevel()
   UINT   marginStopoutType;                       //       372         4     margin stopout type                        = AccountStopoutMode()
   double marginInit;                              //       376         8     initial margin requirement (in units)      = MarketInfo(MODE_MARGININIT)
   double marginMaintenance;                       //       384         8     maintainance margin requirement (in units) = MarketInfo(MODE_MARGINMAINTENANCE)
   double marginHedged;                            //       392         8     hedged margin requirement (in units)       = MarketInfo(MODE_MARGINHEDGED)
   double marginDivider;                           //       400         8     leverage calculation                         @see example in struct SYMBOL
   char   marginCurrency[MAX_SYMBOL_LENGTH+1];     //       408        12                                                = AccountCurrency()
   BYTE   reserved_5[4];                           //       420         4     (alignment to the next double)

   // commission calculation parameters            // ----------------------------------------------------------------------------------------------------------------
   double commissionValue;                         //       424         8     commission rate
   UINT   commissionCalculationMode;               //       432         4     0=Money|1=Pips|2=Percent                     @see COMMISSION_MODE_*
   UINT   commissionType;                          //       436         4     0=RoundTurn|1=PerDeal                        @see COMMISSION_TYPE_*

   // later additions                              // ----------------------------------------------------------------------------------------------------------------
   UINT   firstBar;                                //       440         4     bar number/index??? of first bar (w/o prolog) or 0 for first bar
   UINT   lastBar;                                 //       444         4     bar number/index??? of last bar (w/o prolog) or 0 for last bar
   UINT   startPeriodM1;                           //       448         4     bar index where modeling started using M1 bars
   UINT   startPeriodM5;                           //       452         4     bar index where modeling started using M5 bars
   UINT   startPeriodM15;                          //       456         4     bar index where modeling started using M15 bars
   UINT   startPeriodM30;                          //       460         4     bar index where modeling started using M30 bars
   UINT   startPeriodH1;                           //       464         4     bar index where modeling started using H1 bars
   UINT   startPeriodH4;                           //       468         4     bar index where modeling started using H4 bars
   UINT   testerSettingFrom;                       //       472         4     begin date from tester settings
   UINT   testerSettingTo;                         //       476         4     end date from tester settings
   UINT   freezeDistance;                          //       480         4     order freeze level in points               = MarketInfo(MODE_FREEZELEVEL)
   UINT   modelErrors;                             //       484         4     number of errors during model generation (FIX ERRORS SHOWING UP HERE BEFORE TESTING)
   BYTE   reserved_6[240];                         //       488       240     unused
};

*/

// FXTHeader ...
//
type FXTHeader struct {
	//Header layout--------------------- offset --- size --- description ------------------------------------------------------------------------
	Version      uint32    //  				  0        4     header version:
	Description  [64]byte  //  				  4       64     copyright/description (szchar)
	ServerName   [128]byte //  				 68      128     account server name (szchar)
	Symbol       [12]byte  //  				196       12     symbol (szchar)
	Period       uint32    //  				208        4     timeframe in minutes
	ModelType    uint32    //  				212        4     0=EveryTick|1=ControlPoints|2=BarOpen
	ModeledBars  uint32    //  				216        4     number of modeled bars      (w/o prolog)
	FirstBarTime uint32    //  				220        4     bar open time of first tick (w/o prolog)
	LastBarTime  uint32    //  				224        4     bar open time of last tick  (w/o prolog)
	Reserved1    [4]byte   //  				228        4     (alignment to the next double)
	ModelQuality float64   //  				232        8     max. 99.9

	// common parameters-----------------------------------------------------------------------------------------------------------------------
	BaseCurrency [12]byte // 				240       12     base currency (szchar)                     = StringLeft(symbol, 3)
	Spread       uint32   // 				252        4     spread in points: 0=zero spread            = MarketInfo(MODE_SPREAD)
	Digits       uint32   // 				256        4     digits                                     = MarketInfo(MODE_DIGITS)
	Reserved2    [4]byte  // 				260        4     (alignment to the next double)
	PointSize    float64  // 				264        8     resolution, ie. 0.0000'1                   = MarketInfo(MODE_POINT)
	MinLotsize   uint32   // 				272        4     min lot size in centi lots (hundredths)    = MarketInfo(MODE_MINLOT)  * 100
	MaxLotsize   uint32   // 				276        4     max lot size in centi lots (hundredths)    = MarketInfo(MODE_MAXLOT)  * 100
	LotStepsize  uint32   // 				280        4     lot stepsize in centi lots (hundredths)    = MarketInfo(MODE_LOTSTEP) * 100
	StopsLevel   uint32   // 				284        4     orders stop distance in points             = MarketInfo(MODE_STOPLEVEL)
	PendingsGTC  uint32   // 				288        4     close pending orders at end of day or GTC
	Reserved3    [4]byte  // 				292        4     (alignment to the next double)

	// profit calculation parameters-------------------------------------------------------------------------------------------------------------
	ContractSize          float64 //  		296        8     ie. 100000                                 = MarketInfo(MODE_LOTSIZE)
	TickValue             float64 //  		304        8     tick value in quote currency (empty)       = MarketInfo(MODE_TICKVALUE)
	TickSize              float64 //  		312        8     tick size (empty)                          = MarketInfo(MODE_TICKSIZE)
	ProfitCalculationMode uint32  //  		320        4     0=Forex|1=CFD|2=Futures                    = MarketInfo(MODE_PROFITCALCMODE)

	// swap calculation parameters -------------------------------------------------------------------------------------------------------------
	SwapEnabled         uint32  //  		324        4     if swaps are to be applied
	SwapCalculationMode int32   //  		328        4     0=Points|1=BaseCurrency|2=Interest|3=MarginCurrency   = MarketInfo(MODE_SWAPTYPE)
	Reserved4           [4]byte //  		332        4     (alignment to the next double)
	SwapLongValue       float64 //  		336        8     long overnight swap value                  = MarketInfo(MODE_SWAPLONG)
	SwapShortValue      float64 //  		344        8     short overnight swap values                = MarketInfo(MODE_SWAPSHORT)
	TripleRolloverDay   uint32  //  		352        4     weekday of triple swaps                    = WEDNESDAY (3)

	// margin calculation parameters -------------------------------------------------------------------------------------------------------------
	AccountLeverage           uint32   // 	356        4     account leverage                           = AccountLeverage()
	FreeMarginCalculationType uint32   // 	360        4     free margin calculation type               = AccountFreeMarginMode()
	MarginCalculationMode     uint32   // 	364        4     margin calculation mode                    = MarketInfo(MODE_MARGINCALCMODE)
	MarginStopoutLevel        uint32   // 	368        4     margin stopout level                       = AccountStopoutLevel()
	MarginStopoutType         uint32   // 	372        4     margin stopout type                        = AccountStopoutMode()
	MarginInit                float64  // 	376        8     initial margin requirement (in units)      = MarketInfo(MODE_MARGININIT)
	MarginMaintenance         float64  // 	384        8     maintainance margin requirement (in units) = MarketInfo(MODE_MARGINMAINTENANCE)
	MarginHedged              float64  // 	392        8     hedged margin requirement (in units)       = MarketInfo(MODE_MARGINHEDGED)
	MarginDivider             float64  // 	400        8     leverage calculation                         @see example in struct SYMBOL
	MarginCurrency            [12]byte // 	408       12                                                = AccountCurrency()
	Reserved5                 [4]byte  // 	420        4     (alignment to the next double)

	// commission calculation parameters ----------------------------------------------------------------------------------------------------------
	CommissionValue           float64 // 	424        8     commission rate
	CommissionCalculationMode int32   // 	432        4     0=Money|1=Pips|2=Percent                     @see COMMISSION_MODE_*
	CommissionType            int32   // 	436        4     0=RoundTurn|1=PerDeal                        @see COMMISSION_TYPE_*

	// later additions-----------------------------------------------------------------------------------------------------------------------------
	FirstBar          uint32    //  		440        4     bar number/index??? of first bar (w/o prolog) or 0 for first bar
	LastBar           uint32    //  		444        4     bar number/index??? of last bar (w/o prolog) or 0 for last bar
	StartPeriodM1     uint32    //  		448        4     bar index where modeling started using M1 bars
	StartPeriodM5     uint32    //  		452        4     bar index where modeling started using M5 bars
	StartPeriodM15    uint32    //  		456        4     bar index where modeling started using M15 bars
	StartPeriodM30    uint32    //  		460        4     bar index where modeling started using M30 bars
	StartPeriodH1     uint32    //  		464        4     bar index where modeling started using H1 bars
	StartPeriodH4     uint32    //  		468        4     bar index where modeling started using H4 bars
	TesterSettingFrom uint32    //  		472        4     begin date from tester settings
	TesterSettingTo   uint32    //  		476        4     end date from tester settings
	FreezeDistance    uint32    //  		480        4     order freeze level in points               = MarketInfo(MODE_FREEZELEVEL)
	ModelErrors       uint32    //  		484        4     number of errors during model generation (FIX ERRORS SHOWING UP HERE BEFORE TESTING
	Reserved6         [240]byte //  		488      240     unused
}

// NewHeader return an predefined FXT header
func NewHeader(version uint32, symbol string, timeframe, spread, model uint32) *FXTHeader {
	h := &FXTHeader{
		Version:      version,
		Period:       timeframe,
		ModelType:    model,
		ModelQuality: 99.9,

		// General parameters.
		Spread:      spread,
		Digits:      5,
		PointSize:   1e-5,
		MinLotsize:  1,
		MaxLotsize:  50000,
		LotStepsize: 1,
		StopsLevel:  10,
		PendingsGTC: 1,

		// Profit Calculation parameters.
		ContractSize:          100000.0,
		TickValue:             0.0,
		TickSize:              0.0,
		ProfitCalculationMode: 0,

		// Swap calculation
		SwapEnabled:         0,
		SwapCalculationMode: 0,
		SwapLongValue:       0.0,
		SwapShortValue:      0.0,
		TripleRolloverDay:   3,

		// Margin calculation.
		AccountLeverage:           100,
		FreeMarginCalculationType: 1,
		MarginCalculationMode:     0,
		MarginStopoutLevel:        30,
		MarginStopoutType:         0,
		MarginInit:                0.0,
		MarginMaintenance:         0.0,
		MarginHedged:              50000.0,
		MarginDivider:             1.25,

		// Commission calculation
		CommissionValue:           0.0,
		CommissionCalculationMode: 1,
		CommissionType:            0,

		//  For internal use
		FirstBar: 1,
	}

	toFixBytes(h.Description[:], "Copyright 2001-2017, MetaQuotes Software Corp.")
	toFixBytes(h.ServerName[:], "Beijing MoreU Tech.")
	toFixBytes(h.Symbol[:], symbol)
	toFixBytes(h.BaseCurrency[:], symbol[:3])
	toFixBytes(h.MarginCurrency[:], symbol[3:])

	return h
}

// FixHeader after all ticks writed to file
func (h *FXTHeader) FixHeader() {

}

func toFixBytes(bs []byte, s string) (int, error) {
	r := strings.NewReader(s)
	return r.Read(bs[:])
}
