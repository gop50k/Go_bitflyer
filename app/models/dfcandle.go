package models

import (
	"sort"
	"time"

	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/config"
	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/tradingalgo"
	talib "github.com/markcheno/go-talib"
)

type DataFrameCandle struct {
	ProductCode   string         `json:"product_code"`
	Duration      time.Duration  `json:"duration"`
	Candles       []Candle       `json:"candles"`
	Smas          []Sma          `json:"smas,omitempty"`
	Emas          []Ema          `json:"emas,omitempty"`
	BBands        *BBands        `json:"bbands,omitempty"`
	IchimokuCloud *IchimokuCloud `json:"ichimoku,omitempty"`
	Rsi           *Rsi           `json:"rsi,omitempty"`
	Macd          *Macd          `json:"macd,omitempty"`
	Hvs           []Hv           `json:"hvs,omitempty"`
	Events        *SignalEvents  `json:"events,omitempty"`
}

// Sma 単純移動平均線を取得するStructを作成
type Sma struct {
	Period int       `json:"period,omitempty"`
	Values []float64 `json:"values,omitempty"`
}

// Ema 指数平滑移動平均線を取得するStructを作成
type Ema struct {
	Period int       `json:"period,omitempty"`
	Values []float64 `json:"values,omitempty"`
}

// BBands ボリンジャーバンドを取得するStructを作成
type BBands struct {
	N    int       `json:"n,omitempty"`
	K    float64   `json:"k,omitempty"`
	Up   []float64 `json:"up,omitempty"`
	Mid  []float64 `json:"mid,omitempty"`
	Down []float64 `json:"down,omitempty"`
}

// IchimokuCloud  一目均衡表を取得するStructを作成
type IchimokuCloud struct {
	Tenkan  []float64 `json:"tenkan,omitempty"`
	Kijun   []float64 `json:"kijun,omitempty"`
	SenkouA []float64 `json:"senkoua,omitempty"`
	SenkouB []float64 `json:"senkoub,omitempty"`
	Chikou  []float64 `json:"chikou,omitempty"`
}

// RSI を取得するStructを作成
type Rsi struct {
	Period int       `json:"period,omitempty"`
	Values []float64 `json:"values,omitempty"`
}

// MACD を取得するStructを作成
type Macd struct {
	FastPeriod   int       `json:"fast_period,omitempty"`
	SlowPeriod   int       `json:"slow_period,omitempty"`
	SignalPeriod int       `json:"signal_period,omitempty"`
	Macd         []float64 `json:"macd,omitempty"`
	MacdSignal   []float64 `json:"macd_signal,omitempty"`
	MacdHist     []float64 `json:"macd_hist,omitempty"`
}

// HVS を取得するStructを作成
type Hv struct {
	Period int       `json:"period,omitempty"`
	Values []float64 `json:"values,omitempty"`
}

// DataFrameCandle Srtruct のCandleの Time だけ返すfunction
func (df *DataFrameCandle) Times() []time.Time {
	s := make([]time.Time, len(df.Candles))
	for i, candle := range df.Candles {
		s[i] = candle.Time
	}
	return s
}

// DataFrameCandle Srtruct のCandleの Opens だけ返すfunction
func (df *DataFrameCandle) Opens() []float64 {
	s := make([]float64, len(df.Candles))
	for i, candle := range df.Candles {
		s[i] = candle.Open
	}
	return s
}

// DataFrameCandle Srtruct のCandleの Closes だけ返すfunction
func (df *DataFrameCandle) Closes() []float64 {
	s := make([]float64, len(df.Candles))
	for i, candle := range df.Candles {
		s[i] = candle.Close
	}
	return s
}

// DataFrameCandle Srtruct のCandleの Highs だけ返すfunction
func (df *DataFrameCandle) Highs() []float64 {
	s := make([]float64, len(df.Candles))
	for i, candle := range df.Candles {
		s[i] = candle.High
	}
	return s
}

// DataFrameCandle Srtruct のCandleの Low だけ返すfunction
func (df *DataFrameCandle) Low() []float64 {
	s := make([]float64, len(df.Candles))
	for i, candle := range df.Candles {
		s[i] = candle.Low
	}
	return s
}

// DataFrameCandle Srtruct のCandleの Volume だけ返すfunction
func (df *DataFrameCandle) Volume() []float64 {
	s := make([]float64, len(df.Candles))
	for i, candle := range df.Candles {
		s[i] = candle.Volume
	}
	return s
}

// Sma 単純移動平均値の計算を行う function
func (df *DataFrameCandle) AddSma(period int) bool {
	if len(df.Candles) > period {
		df.Smas = append(df.Smas, Sma{
			Period: period,
			Values: talib.Sma(df.Closes(), period),
		})
		return true
	}
	// period が candleを超えていたら計算出来ないのでfalseを返す
	return false
}

// Ema 指数平滑移動平均線の計算を行う function
func (df *DataFrameCandle) AddEma(period int) bool {
	if len(df.Candles) > period {
		df.Emas = append(df.Emas, Ema{
			Period: period,
			Values: talib.Ema(df.Closes(), period),
		})
		return true
	}
	// period が candleを超えていたら計算出来ないのでfalseを返す
	return false
}

// BBands ボリンジャーバンドの計算を行う function
func (df *DataFrameCandle) AddBBands(n int, k float64) bool {
	if n <= len(df.Closes()) {
		// talib の BBands関数でup, mid, downを取得してBBandsに格納する
		up, mid, down := talib.BBands(df.Closes(), n, k, k, 0)
		df.BBands = &BBands{
			N:    n,
			K:    k,
			Up:   up,
			Mid:  mid,
			Down: down,
		}
		return true
	}
	return false
}

// Ichimoku   一目均衡表を作成する function
func (df *DataFrameCandle) AddIchimoku() bool {
	tenkanN := 9
	if len(df.Closes()) >= tenkanN {
		// talib に IchimokuCloud　が存在しないので　tradingalgo　で自分で作ったfunctionを参照しに行く
		tenkan, kijun, senkouA, senkouB, chikou := tradingalgo.IchimokuCloud(df.Closes())
		df.IchimokuCloud = &IchimokuCloud{

			Tenkan:  tenkan,
			Kijun:   kijun,
			SenkouA: senkouA,
			SenkouB: senkouB,
			Chikou:  chikou,
		}
		return true
	}
	return false
}

// RSI を作成する function
func (df *DataFrameCandle) AddRsi(period int) bool {
	if len(df.Candles) > period {
		values := talib.Rsi(df.Closes(), period)
		df.Rsi = &Rsi{
			Period: period,
			Values: values,
		}
		return true
	}
	return false
}

// Macd を作成する function
func (df *DataFrameCandle) AddMacd(inFastPeriod, inSlowPeriod, inSignalPeriod int) bool {
	if len(df.Candles) > 1 {
		outMACD, outMACDSignal, outMACDHist := talib.Macd(df.Closes(), inFastPeriod, inSlowPeriod, inSignalPeriod)
		df.Macd = &Macd{
			FastPeriod:   inFastPeriod,
			SlowPeriod:   inSlowPeriod,
			SignalPeriod: inSignalPeriod,
			Macd:         outMACD,
			MacdSignal:   outMACDSignal,
			MacdHist:     outMACDHist,
		}
		return true
	}
	return false
}

// HVS   ヒストリカルボラティリティを作成する function
func (df *DataFrameCandle) AddHv(period int) bool {
	if len(df.Closes()) >= period {
		// talib に Hv　が存在しないので　tradingalgo　で自分で作ったfunctionを参照しに行く
		df.Hvs = append(df.Hvs, Hv{
			Period: period,
			Values: tradingalgo.Hv(df.Closes(), period),
		})
		return true
	}
	return false
}

// 指定時間以降の Events を取得するfunction
func (df *DataFrameCandle) AddEvents(timeTime time.Time) bool {
	signalEvents := GetSignalEventsAfterTime(timeTime)
	if len(signalEvents.Signals) > 0 {
		df.Events = signalEvents
		return true
	}
	return false
}

// EMAによるビットコインの売買を行うfunction
func (df *DataFrameCandle) BackTestEma(period1, period2 int) *SignalEvents {
	lenCandles := len(df.Candles)
	// period未満であれば計算ができない
	if lenCandles <= period1 || lenCandles <= period2 {
		return nil
	}
	SignalEvents := NewSignalEvents()
	// dfから、emaを取得する
	emaValues1 := talib.Ema(df.Closes(), period1)
	emaValues2 := talib.Ema(df.Closes(), period2)

	for i := 1; i < lenCandles; i++ {
		if i < period1 || i < period2 {
			continue
		}
		// ゴールデンクロス(EMAの上昇) が起きたとき、ビットコインを購入する
		if emaValues1[i-1] < emaValues2[i-1] && emaValues1[i] >= emaValues2[i] {
			SignalEvents.Buy(df.ProductCode, df.Candles[i].Time, df.Candles[i].Close, 1.0, false)
		}
		// デッドクロス(EMAの下降)が起きたとき、ビットコインを売却する
		if emaValues1[i-1] > emaValues2[i-1] && emaValues1[i] <= emaValues2[i] {
			SignalEvents.Sell(df.ProductCode, df.Candles[i].Time, df.Candles[i].Close, 1.0, false)
		}

	}
	return SignalEvents
}

// BackTestEmaを最適化するfunction(bestなperiodを見つける)
func (df *DataFrameCandle) OptimizeEma() (performance float64, bestPeriod1 int, bestPeriod2 int) {
	// とりあえずbestperiod1,2 に値を入れる
	bestPeriod1 = 7
	bestPeriod2 = 14

	for period1 := 5; period1 < 50; period1++ {
		for period2 := 12; period2 < 50; period2++ {
			// 順番にbestperiodをバックテストに渡す
			signalEvents := df.BackTestEma(period1, period2)
			if signalEvents == nil {
				continue
			}
			profit := signalEvents.Profit()
			if performance < profit {
				performance = profit
				bestPeriod1 = period1
				bestPeriod2 = period2
			}
		}
	}
	return performance, bestPeriod1, bestPeriod2
}

// ボリンジャーバンドのシミュレーションを行うfunction
func (df *DataFrameCandle) BackTestBb(n int, k float64) *SignalEvents {
	lenCandles := len(df.Candles)

	// lenが短いときはボリンジャーバンドの計算が出来ないので返す
	if lenCandles <= n {
		return nil
	}

	signalEvents := &SignalEvents{}
	bbUp, _, bbDown := talib.BBands(df.Closes(), n, k, k, 0)

	// 値が取得できたらBestボリンジャーバンドを取得する
	for i := 1; i < lenCandles; i++ {
		if i < n {
			continue
		}
		// ボリンジャーバンドの下端より、キャンドルスティックの終値が高くて上昇している時に購入
		if bbDown[i-1] > df.Candles[i-1].Close && bbDown[i] <= df.Candles[i].Close {
			signalEvents.Buy(df.ProductCode, df.Candles[i].Time, df.Candles[i].Close, 1.0, false)
		}
		// ボリンジャーバンドの上端より、キャンドルスティックの終値が低くて下降している時に売却
		if bbUp[i-1] < df.Candles[i-1].Close && bbUp[i] >= df.Candles[i].Close {
			signalEvents.Sell(df.ProductCode, df.Candles[i].Time, df.Candles[i].Close, 1.0, false)
		}
	}
	return signalEvents
}

// 最適なボリンジャーバンドの値を探すfunction
func (df *DataFrameCandle) OptimizeBb() (performance float64, bestN int, bestK float64) {
	bestN = 20
	bestK = 2.0

	for n := 10; n < 20; n++ {
		for k := 1.9; k < 2.1; k += 0.1 {
			signalEvents := df.BackTestBb(n, k)
			if signalEvents == nil {
				continue
			}
			profit := signalEvents.Profit()
			if performance < profit {
				performance = profit
				bestN = n
				bestK = k
			}
		}
	}

	return performance, bestN, bestK
}

// 一目均衡表のシミュレーションを行うfunction
func (df *DataFrameCandle) BackTestIchimoku() *SignalEvents {
	lenCandles := len(df.Candles)

	// lenが短いときは一目均衡表の計算が出来ないので返す
	if lenCandles <= 52 {
		return nil
	}

	var signalEvents SignalEvents
	tenkan, kijun, senkouA, senkouB, chikou := tradingalgo.IchimokuCloud(df.Closes())

	for i := 1; i < lenCandles; i++ {

		// キャンドルスティックが上端が遅行線を上回り、下端が先行線よりも上の場合購入
		if chikou[i-1] < df.Candles[i-1].High && chikou[i] >= df.Candles[i].High &&
			senkouA[i] < df.Candles[i].Low && senkouB[i] < df.Candles[i].Low &&
			tenkan[i] > kijun[i] {
			signalEvents.Buy(df.ProductCode, df.Candles[i].Time, df.Candles[i].Close, 1.0, false)
		}

		// キャンドルスティックが上端が遅行線を下回り、下端が先行線よりも下の場合売却
		if chikou[i-1] > df.Candles[i-1].Low && chikou[i] <= df.Candles[i].Low &&
			senkouA[i] > df.Candles[i].High && senkouB[i] > df.Candles[i].High &&
			tenkan[i] < kijun[i] {
			signalEvents.Sell(df.ProductCode, df.Candles[i].Time, df.Candles[i].Close, 1.0, false)
		}
	}
	return &signalEvents
}

// 最適な一目均衡表の値を探すfunction（デフォルトが最適値）
func (df *DataFrameCandle) OptimizeIchimoku() (performance float64) {
	signalEvents := df.BackTestIchimoku()
	if signalEvents == nil {
		return 0.0
	}
	performance = signalEvents.Profit()
	return performance

}

// MACDの売買シミュレーションを行うfunction
func (df *DataFrameCandle) BackTestMacd(macdFastPeriod, macdSlowPeriod, macdSignalPeriod int) *SignalEvents {
	lenCandles := len(df.Candles)

	if lenCandles <= macdFastPeriod || lenCandles <= macdSlowPeriod || lenCandles <= macdSignalPeriod {
		return nil
	}

	signalEvents := &SignalEvents{}
	// macdFastPeriod などの設定値を受け取る
	outMACD, outMACDSignal, _ := talib.Macd(df.Closes(), macdFastPeriod, macdSlowPeriod, macdSignalPeriod)

	for i := 1; i < lenCandles; i++ {
		// 購入の条件
		if outMACD[i] < 0 &&
			outMACDSignal[i] < 0 &&
			// 前日のMACDがSignalを下回っていて、本日上回っていたら購入
			outMACD[i-1] < outMACDSignal[i-1] &&
			outMACD[i] >= outMACDSignal[i] {
			signalEvents.Sell(df.ProductCode, df.Candles[i].Time, df.Candles[i].Close, 1.0, false)
		}

		// 売却の条件
		if outMACD[i] > 0 &&
			outMACDSignal[i] > 0 &&
			// 前日のMACDがSignalを上回っていて、本日下回っていたら売却
			outMACD[i-1] > outMACDSignal[i-1] &&
			outMACD[i] <= outMACDSignal[i] {
			signalEvents.Sell(df.ProductCode, df.Candles[i].Time, df.Candles[i].Close, 1.0, false)
		}

	}
	return signalEvents
}

// 最適なMACDの値を探すfunction
func (df *DataFrameCandle) OptimizeMacd() (performance float64, bestMacdFastPeriod, bestMacdSlowPeriod, bestMacdSignalPeriod int) {
	bestMacdFastPeriod = 12
	bestMacdSlowPeriod = 26
	bestMacdSignalPeriod = 9

	// 最適なPeriodを探す
	for fastPeriod := 10; fastPeriod < 19; fastPeriod++ {
		for slowPeriod := 20; slowPeriod < 30; slowPeriod++ {
			for signalPeriod := 5; signalPeriod < 15; signalPeriod++ {
				signalEvents := df.BackTestMacd(bestMacdFastPeriod, bestMacdSlowPeriod, bestMacdSignalPeriod)
				if signalEvents == nil {
					continue
				}
				profit := signalEvents.Profit()
				if performance < profit {
					performance = profit
					bestMacdFastPeriod = fastPeriod
					bestMacdSlowPeriod = slowPeriod
					bestMacdSignalPeriod = signalPeriod
				}
			}
		}
	}
	return performance, bestMacdFastPeriod, bestMacdSlowPeriod, bestMacdSignalPeriod
}

// RSIの売買シミュレーションを行うfunction
func (df *DataFrameCandle) BackTestRsi(period int, buyThread, sellThread float64) *SignalEvents {
	lenCandles := len(df.Candles)
	if lenCandles <= period {
		return nil
	}

	signalEvents := NewSignalEvents()
	values := talib.Rsi(df.Closes(), period)
	for i := 1; i < lenCandles; i++ {
		if values[i-1] == 0 || values[i-1] == 100 {
			continue
		}
		// 前日のRSIが下端の線より下で、本日上回れば購入
		if values[i-1] < buyThread && values[i] >= buyThread {
			signalEvents.Buy(df.ProductCode, df.Candles[i].Time, df.Candles[i].Close, 1.0, false)
		}

		// 前日のRSIが上端の線より上で、本日下回れば売却
		if values[i-1] > sellThread && values[i] <= sellThread {
			signalEvents.Sell(df.ProductCode, df.Candles[i].Time, df.Candles[i].Close, 1.0, false)
		}
	}
	return signalEvents
}

// 最適なRSIの値を探すfunction
func (df *DataFrameCandle) OptimizeRsi() (performance float64, bestPeriod int, bestBuyThread, bestSellThread float64) {
	bestPeriod = 14
	// RSIの一般的な購入の指標である30%,売却の指標である70%をデフォルトで入れておく
	bestBuyThread, bestSellThread = 30.0, 70.0

	for period := 5; period < 25; period++ {
		signalEvents := df.BackTestRsi(period, bestBuyThread, bestSellThread)
		if signalEvents == nil {
			continue
		}
		profit := signalEvents.Profit()
		if performance < profit {
			performance = profit

			bestPeriod = period
			bestBuyThread = bestBuyThread
			bestSellThread = bestSellThread
		}
	}
	return performance, bestPeriod, bestBuyThread, bestSellThread

}

// 最適なインディケータを選出するためのStruct
type TradeParams struct {
	EmaEnable        bool
	EmaPeriod1       int
	EmaPeriod2       int
	BbEnable         bool
	BbN              int
	BbK              float64
	IchimokuEnable   bool
	MacdEnable       bool
	MacdFastPeriod   int
	MacdSlowPeriod   int
	MacdSignalPeriod int
	RsiEnable        bool
	RsiPeriod        int
	RsiBuyThread     float64
	RsiSellThread    float64
}

// インディケータのランキングを入れるStruct
type Ranking struct {
	Enable      bool
	Performance float64
}

// 過去に作成したインディケータのパフォーマンスを取得してランキング付けするfunction
func (df *DataFrameCandle) OptimizeParams() *TradeParams {
	emaPerformance, emaPeriod1, emaPeriod2 := df.OptimizeEma()
	bbPerformance, bbN, bbK := df.OptimizeBb()
	macdPerformance, macdFastPeriod, macdSlowPeriod, macdSignalPeriod := df.OptimizeMacd()
	ichimokuPerforamcne := df.OptimizeIchimoku()
	rsiPerformance, rsiPeriod, rsiBuyThread, rsiSellThread := df.OptimizeRsi()

	emaRanking := &Ranking{false, emaPerformance}
	bbRanking := &Ranking{false, bbPerformance}
	macdRanking := &Ranking{false, macdPerformance}
	ichimokuRanking := &Ranking{false, ichimokuPerforamcne}
	rsiRanking := &Ranking{false, rsiPerformance}

	rankings := []*Ranking{emaRanking, bbRanking, macdRanking, ichimokuRanking, rsiRanking}

	// パフォーマンスの大きい順に並び替える
	sort.Slice(rankings, func(i, j int) bool { return rankings[i].Performance > rankings[j].Performance })

	// Configで設定したようにベスト３までランキング付けする
	for i, ranking := range rankings {
		if i >= config.Config.NumRanking {
			break
		}
		// パフォーマンスが出ていればenableとする
		if ranking.Performance > 0 {
			ranking.Enable = true
		}
	}

	tradeParams := &TradeParams{
		EmaEnable:        emaRanking.Enable,
		EmaPeriod1:       emaPeriod1,
		EmaPeriod2:       emaPeriod2,
		BbEnable:         bbRanking.Enable,
		BbN:              bbN,
		BbK:              bbK,
		IchimokuEnable:   ichimokuRanking.Enable,
		MacdEnable:       macdRanking.Enable,
		MacdFastPeriod:   macdFastPeriod,
		MacdSlowPeriod:   macdSlowPeriod,
		MacdSignalPeriod: macdSignalPeriod,
		RsiEnable:        rsiRanking.Enable,
		RsiPeriod:        rsiPeriod,
		RsiBuyThread:     rsiBuyThread,
		RsiSellThread:    rsiSellThread,
	}
	return tradeParams
}
