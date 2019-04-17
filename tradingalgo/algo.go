package tradingalgo

import (
	"math"

	"github.com/markcheno/go-talib"
)

/*
一目均衡表のアルゴリズムを作成する
Tenkan = (9-day high + 9-day low) / 2
Kijun = (26-day high + 26-day low) / 2
Senkou Span A = (Tenkan + Kijun) / 2
Senkou Span B = (52-day high + 52-day low) / 2
Chikou Span = Close plotted 26 days in the past
*/

// Tenkan を作成する為に価格の最安値と最高値を作成する function
func minMax(inReal []float64) (float64, float64) {
	min := inReal[0]
	max := inReal[0]
	for _, price := range inReal {
		if min > price {
			min = price
		}
		if max > price {
			max = price
		}
	}
	return min, max
}

// 最安値を作成する function
func min(x, y int) int {
	if x < y {
		return x
	} else {
		return y
	}
}

// IchimokuCloud　一目均衡表を作成する function
func IchimokuCloud(inReal []float64) ([]float64, []float64, []float64, []float64, []float64) {

	// 一目均衡表の５本の線の空のスライスを作成しておく
	length := len(inReal)
	tenkan := make([]float64, min(9, length))
	kijun := make([]float64, min(26, length))
	senkouA := make([]float64, min(26, length))
	senkouB := make([]float64, min(52, length))
	chikou := make([]float64, min(26, length))

	for i := range inReal {
		// 転換線の作成
		if i >= 9 {
			min, max := minMax(inReal[i-9 : i])
			tenkan = append(tenkan, (min+max)/2)
		}
		// 基準線、先行A、遅行線の作成
		if i >= 26 {
			min, max := minMax(inReal[i-26 : i])
			kijun = append(kijun, (min+max)/2)
			senkouA = append(senkouA, (tenkan[i]+kijun[i])/2)
			chikou = append(chikou, inReal[i-26])
		}
		// 先行Bの作成
		if i >= 52 {
			min, max := minMax(inReal[i-52 : i])
			senkouB = append(senkouB, (min+max)/2)
		}
	}
	return tenkan, kijun, senkouA, senkouB, chikou
}

// Hv　ヒストリカルボラティリティを作成する function
func Hv(inReal []float64, inTimePeriod int) []float64 {
	change := make([]float64, 0)
	for i := range inReal {
		if i == 0 {
			continue
		}
		// 日毎のログを格納する
		dayChange := math.Log(
			float64(inReal[i]) / float64(inReal[i-1]))
		change = append(change, dayChange)
	}
	// 前日と当日のログを計算して標準偏差を返す
	return talib.StdDev(change, inTimePeriod, math.Sqrt(1)*100)
}
