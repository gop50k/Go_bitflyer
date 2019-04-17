package controllers

import (
	"log"
	"strings"
	"time"

	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/app/models"
	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/bitflyer"
	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/config"
	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/tradingalgo"
	talib "github.com/markcheno/go-talib"
	"golang.org/x/sync/semaphore"
)

type AI struct {
	API                  *bitflyer.APIClient
	ProductCode          string
	CurrencyCode         string
	CoinCode             string
	UsePercent           float64
	MinuteToExpires      int
	Duration             time.Duration
	PastPeriod           int
	SignalEvents         *models.SignalEvents
	OptimizedTradeParams *models.TradeParams
	TradeSemaphore       *semaphore.Weighted
	StopLimit            float64
	StopLimitPercent     float64
	BackTest             bool
	StartTrade           time.Time
}

// グローバルで宣言
var Ai *AI

func NewAI(productCode string, duration time.Duration, pastPeriod int, UsePercent, stopLimitPercent float64, backTest bool) *AI {
	apiClient := bitflyer.New(config.Config.APIKey, config.Config.APISecret)
	var signalEvents *models.SignalEvents
	// バックテストの場合
	if backTest {
		signalEvents = models.NewSignalEvents()
	} else {
		// 再起動などを行なった際に、購入か売却かを判断する
		signalEvents = models.GetSignalEventsByCount(1)
	}
	codes := strings.Split(productCode, "_")

	// グローバルで宣言したAIに格納する
	Ai = &AI{
		API:              apiClient,
		ProductCode:      productCode,
		CoinCode:         codes[0],
		CurrencyCode:     codes[1],
		UsePercent:       UsePercent,
		MinuteToExpires:  1,
		PastPeriod:       pastPeriod,
		Duration:         duration,
		SignalEvents:     signalEvents,
		TradeSemaphore:   semaphore.NewWeighted(1),
		BackTest:         backTest,
		StartTrade:       time.Now(),
		StopLimitPercent: stopLimitPercent,
	}
	// インディケータの最適値を入れる
	Ai.UpdateOptimizeParams()
	return Ai
}

func (ai *AI) UpdateOptimizeParams() {
	df, _ := models.GetAllCandle(ai.ProductCode, ai.Duration, ai.PastPeriod)
	// インディケータの最適化した結果を ai に格納する
	ai.OptimizedTradeParams = df.OptimizeParams()
	log.Printf("optimized_trade_params=%+v", ai.OptimizedTradeParams)
}

// AI で購入を行う function
func (ai *AI) Buy(candle models.Candle) (childOrderAcceptanceID string, isOrderCompleted bool) {
	// アカウントを持っていない為、バックテストで実行
	if ai.BackTest {
		couldBuy := ai.SignalEvents.Buy(ai.ProductCode, candle.Time, candle.Close, 1.0, false)
		return "", couldBuy
	}

	// アカウントを持っている場合、実際の購入を実行できる
	return childOrderAcceptanceID, isOrderCompleted
}

// AI で売却を行う function
func (ai *AI) Sell(candle models.Candle) (childOrderAcceptanceID string, isOrderCompleted bool) {
	// アカウントを持っていない為、バックテストで実行
	if ai.BackTest {
		couldSell := ai.SignalEvents.Sell(ai.ProductCode, candle.Time, candle.Close, 1.0, false)
		return "", couldSell
	}

	// アカウントを持っている場合、実際の売却を実行できる
	return childOrderAcceptanceID, isOrderCompleted
}

// トレードを行う function
func (ai *AI) Trade() {
	isAcquire := ai.TradeSemaphore.TryAcquire(1)
	if !isAcquire {
		log.Println("Could not get trade lock")
		return
	}
	defer ai.TradeSemaphore.Release(1)
	params := ai.OptimizedTradeParams
	df, _ := models.GetAllCandle(ai.ProductCode, ai.Duration, ai.PastPeriod)
	lenCandles := len(df.Candles)

	// 最適化された値を取得して、Enableであれば使用
	var emaValues1 []float64
	var emaValues2 []float64
	if params.EmaEnable {
		emaValues1 = talib.Ema(df.Closes(), params.EmaPeriod1)
		emaValues2 = talib.Ema(df.Closes(), params.EmaPeriod2)
	}

	var bbUp []float64
	var bbDown []float64
	if params.BbEnable {
		bbUp, _, bbDown = talib.BBands(df.Closes(), params.BbN, params.BbK, params.BbK, 0)
	}

	var tenkan, kijun, senkouA, senkouB, chikou []float64
	if params.IchimokuEnable {
		tenkan, kijun, senkouA, senkouB, chikou = tradingalgo.IchimokuCloud(df.Closes())
	}

	var outMACD, outMACDSignal []float64
	if params.MacdEnable {
		outMACD, outMACDSignal, _ = talib.Macd(df.Closes(), params.MacdFastPeriod, params.MacdSlowPeriod, params.MacdSignalPeriod)
	}

	var rsiValues []float64
	if params.RsiEnable {
		rsiValues = talib.Rsi(df.Closes(), params.RsiPeriod)
	}

	for i := 1; i < lenCandles; i++ {
		// 売買を行う基準となる buyPoint, sellPointを宣言
		buyPoint, sellPoint := 0, 0
		// ゴールデンクロスが計算できるか判定
		if params.EmaEnable && params.EmaPeriod1 <= i && params.EmaPeriod2 <= i {
			if emaValues1[i-1] < emaValues2[i-1] && emaValues1[i] >= emaValues2[i] {
				buyPoint++
			}

			if emaValues1[i-1] > emaValues2[i-1] && emaValues1[i] <= emaValues2[i] {
				sellPoint++
			}
		}

		// ボリンジャーバンドが計算できるか判定
		if params.BbEnable && params.BbN <= i {
			if bbDown[i-1] > df.Candles[i-1].Close && bbDown[i] <= df.Candles[i].Close {
				buyPoint++
			}

			if bbUp[i-1] < df.Candles[i-1].Close && bbUp[i] >= df.Candles[i].Close {
				sellPoint++
			}
		}

		// Macdが計算できるか判定
		if params.MacdEnable {
			if outMACD[i] < 0 && outMACDSignal[i] < 0 && outMACD[i-1] < outMACDSignal[i-1] && outMACD[i] >= outMACDSignal[i] {
				buyPoint++
			}

			if outMACD[i] > 0 && outMACDSignal[i] > 0 && outMACD[i-1] > outMACDSignal[i-1] && outMACD[i] <= outMACDSignal[i] {
				sellPoint++
			}
		}

		// 一目均衡表が計算できるか判定
		if params.IchimokuEnable {
			if chikou[i-1] < df.Candles[i-1].High && chikou[i] >= df.Candles[i].High &&
				senkouA[i] < df.Candles[i].Low && senkouB[i] < df.Candles[i].Low &&
				tenkan[i] > kijun[i] {
				buyPoint++
			}

			if chikou[i-1] > df.Candles[i-1].Low && chikou[i] <= df.Candles[i].Low &&
				senkouA[i] > df.Candles[i].High && senkouB[i] > df.Candles[i].High &&
				tenkan[i] < kijun[i] {
				sellPoint++
			}
		}

		// RSIが計算できるか判定
		if params.RsiEnable && rsiValues[i-1] != 0 && rsiValues[i-1] != 100 {
			if rsiValues[i-1] < params.RsiBuyThread && rsiValues[i] >= params.RsiBuyThread {
				buyPoint++
			}

			if rsiValues[i-1] > params.RsiSellThread && rsiValues[i] <= params.RsiSellThread {
				sellPoint++
			}
		}

		// buyPointが０以上であれば購入（最適化されたインディケータが buyPoint++ すれば購入）
		if buyPoint > 0 {
			_, isOrderCompleted := ai.Buy(df.Candles[i])
			if !isOrderCompleted {
				continue
			}
			// 終値に設定したパーセンテージを掛けた値に達した場合終了
			ai.StopLimit = df.Candles[i].Close * ai.StopLimitPercent
		}

		// 終値が StopLimit より下降した場合、もしくは SellPoint++ した場合売却
		if sellPoint > 0 || ai.StopLimit > df.Candles[i].Close {
			_, isOrderCompleted := ai.Sell(df.Candles[i])
			if !isOrderCompleted {
				continue
			}
			ai.StopLimit = 0.0
			ai.UpdateOptimizeParams()
		}
	}
}
