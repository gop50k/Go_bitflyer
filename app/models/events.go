package models

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/config"
)

// ビットコインを売買したイベントをDBに格納していく

// DBのフィールドを入れる Struct
type SignalEvent struct {
	Time        time.Time `json:"time"`
	ProductCode string    `json:"product_code"`
	Side        string    `json:"side"`
	Price       float64   `json:"price"`
	Size        float64   `json:"size"`
}

// tableNameSignalEvents にデータを格納するfunction
func (s *SignalEvent) Save() bool {
	cmd := fmt.Sprintf("INSERT INTO %s (time, product_code, side, price, size) VALUES (?, ?, ?, ?, ?)", tableNameSignalEvents)
	_, err := DbConnection.Exec(cmd, s.Time.Format(time.RFC3339), s.ProductCode, s.Side, s.Price, s.Size)
	if err != nil {
		// 間違って同じ時間で購入してしまう時のエラーハンドリング
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			log.Panicln(err)
			return true
		}
		return false
	}
	return true
}

// 売買の記録を順番通りに格納するスライスを定義する
type SignalEvents struct {
	Signals []SignalEvent `json:"signals,omitempty"`
}

func NewSignalEvents() *SignalEvents {
	return &SignalEvents{}
}

// 損益を計算する

// もしDBにイベント情報があったら順番を入れ替えて tableNameSignalEvents　に格納するfunction
func GetSignalEventsByCount(loadEvents int) *SignalEvents {
	cmd := fmt.Sprintf(`SELECT * FROM (
		SELECT time, product_code, side, price, size FROM %s WHERE product_code = ? ORDER BY time DESC LIMIT ?)
		ORDER BY time ASC;`, tableNameSignalEvents)
	rows, err := DbConnection.Query(cmd, config.Config.ProductCode, loadEvents)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var signalEvents SignalEvents
	for rows.Next() {
		var signalEvent SignalEvent
		rows.Scan(&signalEvent.Time, &signalEvent.ProductCode, &signalEvent.Side, &signalEvent.Price, &signalEvent.Size)
		signalEvents.Signals = append(signalEvents.Signals, signalEvent)
	}
	err = rows.Err()
	if err != nil {
		return nil
	}
	return &signalEvents
}

// 指定した時間以降の売買のデータを取得して tableNameSignalEvents　に格納するfunction
func GetSignalEventsAfterTime(timeTime time.Time) *SignalEvents {
	cmd := fmt.Sprintf(`SELECT * FROM (
		SELECT time, product_code, side, price, size FROM %s
		WHERE DATETIME(time) >= DATETIME(?)
		ORDER BY time DESC
	) ORDER BY time ASC;`, tableNameSignalEvents)
	rows, err := DbConnection.Query(cmd, timeTime.Format(time.RFC3339))
	if err != nil {
		return nil
	}
	defer rows.Close()

	var signalEvents SignalEvents
	for rows.Next() {
		var signalEvent SignalEvent
		rows.Scan(&signalEvent.Time, &signalEvent.ProductCode, &signalEvent.Side, &signalEvent.Price, &signalEvent.Size)
		signalEvents.Signals = append(signalEvents.Signals, signalEvent)
	}
	return &signalEvents
}

// 実際に購入できるか判定する function
func (s *SignalEvents) CanBuy(time time.Time) bool {
	// SignalEvents の中にはデータがあるか判定
	lenSignals := len(s.Signals)
	if lenSignals == 0 {
		return true
	}

	//  最後のデータが売りであれば購入可能、現在時間よりも最後の購入時間が前であれば購入可能
	lastSignal := s.Signals[lenSignals-1]
	if lastSignal.Side == "SELL" && lastSignal.Time.Before(time) {
		return true
	}
	return false
}

// 実際に売却できるか判定する function
func (s *SignalEvents) CanSell(time time.Time) bool {
	// SignalEvents の中にはデータがあるか判定
	lenSignals := len(s.Signals)
	if lenSignals == 0 {
		return true
	}

	//  最後のデータが売りであれば売却可能、現在時間よりも最後の売却時間が前であれば購入可能
	lastSignal := s.Signals[lenSignals-1]
	if lastSignal.Side == "BUY" && lastSignal.Time.Before(time) {
		return true
	}
	return false
}

// 購入を行う function
func (s *SignalEvents) Buy(ProductCode string, time time.Time, price, size float64, save bool) bool {

	// 購入可能か判定
	if !s.CanBuy(time) {
		return false
	}
	signalEvent := SignalEvent{
		ProductCode: ProductCode,
		Time:        time,
		Side:        "BUY",
		Price:       price,
		Size:        size,
	}
	// シミュレーションとして、実際の売買を行わない時はDBに格納しない
	if save {
		signalEvent.Save()
	}
	s.Signals = append(s.Signals, signalEvent)
	return true
}

// 売却を行う function
func (s *SignalEvents) Sell(ProductCode string, time time.Time, price, size float64, save bool) bool {

	// 売却可能か判定
	if !s.CanSell(time) {
		return false
	}
	signalEvent := SignalEvent{
		ProductCode: ProductCode,
		Time:        time,
		Side:        "SELL",
		Price:       price,
		Size:        size,
	}
	// シミュレーションとして、実際の売買を行わない時はDBに格納しない
	if save {
		signalEvent.Save()
	}
	s.Signals = append(s.Signals, signalEvent)
	return true
}

// 売買の profit(利益)を計算する function
func (s *SignalEvents) Profit() float64 {
	total := 0.0
	beforeSell := 0.0
	// 購入したがまだ売却していない状況の変数
	isHolding := false
	// 始めがSELLだと計算ができない
	for i, signalEvent := range s.Signals {
		if i == 0 && signalEvent.Side == "SELL" {
			continue
		}
		// 購入したので資金をマイナスにする
		if signalEvent.Side == "BUY" {
			total -= signalEvent.Price * signalEvent.Size
			isHolding = true
		}
		// 売却したので資金をプラスにする
		if signalEvent.Side == "SELL" {
			total -= signalEvent.Price * signalEvent.Size
			isHolding = false
			beforeSell = total
		}
		if isHolding == true {
			return beforeSell
		}
	}
	// 総資産を返す
	return total
}

// profit をjsonにして渡すfunction
func (s SignalEvents) MarshalJSON() ([]byte, error) {
	value, err := json.Marshal(&struct {
		Signals []SignalEvent `json:"signals,omitempty"`
		Profit  float64       `json:"profit,omitempty"`
	}{
		Signals: s.Signals,
		Profit:  s.Profit(),
	})
	if err != nil {
		return nil, err
	}
	return value, err
}

// シミュレーションを行う際に必要な、DBに格納されていないデータを取得するfunction
func (s *SignalEvents) CollectAfter(time time.Time) *SignalEvents {
	for i, signal := range s.Signals {
		if time.After(signal.Time) {
			continue
		}
		return &SignalEvents{Signals: s.Signals[i:]}
	}
	return nil
}
