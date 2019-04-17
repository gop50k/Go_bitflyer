package controllers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/app/models"
	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/config"
)

// chart.html を読み込むファイル
var templates = template.Must(template.ParseFiles("backend-2-go-fintech/Section20/app/views/chart.html"))

func viewChartHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "chart.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ajaxを用いて非同期でデータを取得できるようにする
type JSONError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// エラーメッセージとエラーコードを返すfunction
func APIError(w http.ResponseWriter, errMessage string, code int) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	jsonError, err := json.Marshal(JSONError{Error: errMessage, Code: code})
	if err != nil {
		log.Fatal(err)
	}
	w.Write(jsonError)
}

var apiValidPath = regexp.MustCompile("^/api/candle/$")

// apiValidPathにマッチングする物があるか調べるfunction
func apiMakeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := apiValidPath.FindStringSubmatch(r.URL.Path)
		if len(m) == 0 {
			APIError(w, "Not found", http.StatusNotFound)
		}
		fn(w, r)
	}
}

// Jsonにして値を返す function
func apiCandleHandler(w http.ResponseWriter, r *http.Request) {
	productCode := r.URL.Query().Get("product_code")
	if productCode == "" {
		APIError(w, "No product_code param", http.StatusBadRequest)
		return
	}
	// 各パラメータの限界値を設定
	strLimit := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(strLimit)
	if strLimit == "" || err != nil || limit < 0 || limit > 1000 {
		limit = 1000
	}
	duration := r.URL.Query().Get("duration")
	if duration == "" {
		duration = "1m"
	}
	durationTime := config.Config.Durations[duration]

	df, _ := models.GetAllCandle(productCode, durationTime, limit)

	// フロントエンドからSmaが来たらdfに追加する
	sma := r.URL.Query().Get("sma")
	if sma != "" {
		strSmaPeriod1 := r.URL.Query().Get("smaPeriod1")
		strSmaPeriod2 := r.URL.Query().Get("smaPeriod2")
		strSmaPeriod3 := r.URL.Query().Get("smaPeriod3")
		period1, err := strconv.Atoi(strSmaPeriod1)
		// 変なデータが入ってきた時の為にデフォルト値を入れておく
		if strSmaPeriod1 == "" || err != nil || period1 < 0 {
			period1 = 7
		}

		period2, err := strconv.Atoi(strSmaPeriod2)
		if strSmaPeriod2 == "" || err != nil || period2 < 0 {
			period2 = 14
		}

		period3, err := strconv.Atoi(strSmaPeriod3)
		if strSmaPeriod3 == "" || err != nil || period3 < 0 {
			period3 = 50
		}
		df.AddSma(period1)
		df.AddSma(period2)
		df.AddSma(period3)
	}

	// フロントエンドからEmaが来たらdfに追加する
	ema := r.URL.Query().Get("ema")
	if ema != "" {
		strEmaPeriod1 := r.URL.Query().Get("emaPeriod1")
		strEmaPeriod2 := r.URL.Query().Get("emaPeriod2")
		strEmaPeriod3 := r.URL.Query().Get("emaPeriod3")
		period1, err := strconv.Atoi(strEmaPeriod1)
		// 変なデータが入ってきた時の為にデフォルト値を入れておく
		if strEmaPeriod1 == "" || err != nil || period1 < 0 {
			period1 = 7
		}

		period2, err := strconv.Atoi(strEmaPeriod2)
		if strEmaPeriod2 == "" || err != nil || period2 < 0 {
			period2 = 14
		}

		period3, err := strconv.Atoi(strEmaPeriod3)
		if strEmaPeriod3 == "" || err != nil || period3 < 0 {
			period3 = 50
		}
		df.AddEma(period1)
		df.AddEma(period2)
		df.AddEma(period3)
	}

	// フロントエンドからbbandsが来たらdfに追加する
	bbands := r.URL.Query().Get("bbands")
	if bbands != "" {
		strN := r.URL.Query().Get("bbandsN")
		strK := r.URL.Query().Get("bbandsK")
		n, err := strconv.Atoi(strN)
		// 変なデータが入ってきた時の為にデフォルト値を入れておく
		if strN == "" || err != nil || n < 0 {
			n = 20
		}

		k, err := strconv.Atoi(strK)
		if strK == "" || err != nil || k < 0 {
			k = 2
		}
		df.AddBBands(n, float64(k))
	}

	// フロントエンドからichimokuが来たらdfに追加する
	ichimoku := r.URL.Query().Get("ichimoku")
	if ichimoku != "" {
		df.AddIchimoku()
	}

	// フロントエンドからrsiが来たらdfに追加する
	rsi := r.URL.Query().Get("rsi")
	if rsi != "" {
		strPeriod := r.URL.Query().Get("rsiPeriod")
		period, err := strconv.Atoi(strPeriod)
		if strPeriod == "" || err != nil || period < 0 {
			period = 14
		}
		df.AddRsi(period)
	}

	// フロントエンドからmacdが来たらdfに追加する
	macd := r.URL.Query().Get("macd")
	if macd != "" {
		strPeriod1 := r.URL.Query().Get("macdPeriod1")
		strPeriod2 := r.URL.Query().Get("macdPeriod2")
		strPeriod3 := r.URL.Query().Get("macdPeriod3")
		period1, err := strconv.Atoi(strPeriod1)
		if strPeriod1 == "" || err != nil || period1 < 0 {
			period1 = 12
		}
		period2, err := strconv.Atoi(strPeriod2)
		if strPeriod2 == "" || err != nil || period2 < 0 {
			period2 = 26
		}
		period3, err := strconv.Atoi(strPeriod3)
		if strPeriod3 == "" || err != nil || period3 < 0 {
			period3 = 9
		}
		df.AddMacd(period1, period2, period3)
	}

	// フロントエンドからhvが来たらdfに追加する
	hv := r.URL.Query().Get("hv")
	if hv != "" {
		strPeriod1 := r.URL.Query().Get("hvPeriod1")
		strPeriod2 := r.URL.Query().Get("hvPeriod2")
		strPeriod3 := r.URL.Query().Get("hvPeriod3")
		period1, err := strconv.Atoi(strPeriod1)
		if strPeriod1 == "" || err != nil || period1 < 0 {
			period1 = 21
		}
		period2, err := strconv.Atoi(strPeriod2)
		if strPeriod2 == "" || err != nil || period2 < 0 {
			period2 = 63
		}
		period3, err := strconv.Atoi(strPeriod3)
		if strPeriod3 == "" || err != nil || period3 < 0 {
			period3 = 252
		}
		df.AddHv(period1)
		df.AddHv(period2)
		df.AddHv(period3)
	}

	events := r.URL.Query().Get("events")
	if events != "" {

		// バックテストの場合
		if config.Config.BackTest {
			df.Events = Ai.SignalEvents.CollectAfter(df.Candles[0].Time)
		} else {
			// バックテストではない場合
			firstTime := df.Candles[0].Time
			df.AddEvents(firstTime)
		}
	}

	// productCode, durationTime, limit が格納されたJsonを変換する
	js, err := json.Marshal(df)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// Handler の登録、サーバーの立ち上げを行うfunction
func StartWebServer() error {
	http.HandleFunc("/api/candle/", apiMakeHandler(apiCandleHandler))
	http.HandleFunc("/chart/", viewChartHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Config.Port), nil)
}
