package models

import (
	"fmt"
	"time"

	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/bitflyer"
)

// Candle キャンドルスティックのモデルを作成
type Candle struct {
	ProductCode string        `json:"product_code"`
	Duration    time.Duration `json:"duration"`
	Time        time.Time     `json:"time"`
	Open        float64       `json:"open"`
	Close       float64       `json:"close"`
	High        float64       `json:"high"`
	Low         float64       `json:"low"`
	Volume      float64       `json:"volume"`
}

// NewCandle Candleを作成するfunction
func NewCandle(productCode string, duration time.Duration, timeDate time.Time, open, close, high, low, volume float64) *Candle {
	return &Candle{
		productCode,
		duration,
		timeDate,
		open,
		close,
		high,
		low,
		volume,
	}
}

// TableName テーブルネームを取ってくるfunction
func (c *Candle) TableName() string {
	return GetCandleTableName(c.ProductCode, c.Duration)
}

// Create テーブル作成する(キャンドルスティックを作成する)function
func (c *Candle) Create() error {
	cmd := fmt.Sprintf("INSERT INTO %s (time, open, close, high, low, volume) VALUES (?, ?, ?, ?, ?, ?)", c.TableName())
	_, err := DbConnection.Exec(cmd, c.Time.Format(time.RFC3339), c.Open, c.Close, c.High, c.Low, c.Volume)
	if err != nil {
		return err
	}
	return err
}

// Save テーブルをアップデートする(キャンドルスティックを更新する)function
func (c *Candle) Save() error {
	cmd := fmt.Sprintf("UPDATE %s SET open = ?, close = ?, high = ?, low = ?, volume = ? WHERE time = ?", c.TableName())
	_, err := DbConnection.Exec(cmd, c.Open, c.Close, c.High, c.Low, c.Volume, c.Time.Format(time.RFC3339))
	if err != nil {
		return err
	}
	return err
}

// GetCandle Selectするfunction
func GetCandle(productCode string, duration time.Duration, dateTime time.Time) *Candle {
	tableName := GetCandleTableName(productCode, duration)
	cmd := fmt.Sprintf("SELECT time, open, close, high, low, volume FROM  %s WHERE time = ?", tableName)
	row := DbConnection.QueryRow(cmd, dateTime.Format(time.RFC3339))
	var candle Candle
	err := row.Scan(&candle.Time, &candle.Open, &candle.Close, &candle.High, &candle.Low, &candle.Volume)
	if err != nil {
		return nil
	}
	return NewCandle(productCode, duration, candle.Time, candle.Open, candle.Close, candle.High, candle.Low, candle.Volume)
}

// CreateCandleWithDuration Tickerで取得した情報をもとにDBに書き込みキャンドルスティックを作成するfunction
func CreateCandleWithDuration(ticker bitflyer.Ticker, productCode string, duration time.Duration) bool {
	// 直近のキャンドルを定義
	currentCandle := GetCandle(productCode, duration, ticker.TruncateDateTime(duration))
	// 金額の中間地点を取得
	price := ticker.GetMidPrice()
	// currentCandle が存在しない場合、DBにcurrentCandleを作成する
	if currentCandle == nil {
		candle := NewCandle(productCode, duration, ticker.TruncateDateTime(duration),
			price, price, price, price, ticker.Volume)
		candle.Create()
		return true
	}
	// 存在するキャンドルと取得したキャンドルの価格に差異があった場合、書き換える
	if currentCandle.High <= price {
		currentCandle.High = price
	} else if currentCandle.Low >= price {
		currentCandle.Low = price
	}
	// 出来高を蓄積
	currentCandle.Volume += ticker.Volume
	// マーケットがクローズした時の金額を格納
	currentCandle.Close = price
	currentCandle.Save()
	return false
}

// 最新のTicker情報を取得してデータフレームに返すfunction
func GetAllCandle(productCode string, duration time.Duration, limit int) (dfCandle *DataFrameCandle, err error) {
	tableName := GetCandleTableName(productCode, duration)
	cmd := fmt.Sprintf(`SELECT * FROM (
	SELECT time, open, close, high, low, volume FROM %s ORDER BY time DESC LIMIT ?
) ORDER BY time ASC;`, tableName)
	rows, err := DbConnection.Query(cmd, limit)
	if err != nil {
		return
	}
	defer rows.Close()

	dfCandle = &DataFrameCandle{}
	dfCandle.ProductCode = productCode
	dfCandle.Duration = duration
	for rows.Next() {
		var candle Candle
		candle.ProductCode = productCode
		candle.Duration = duration
		rows.Scan(&candle.Time, &candle.Open, &candle.Close, &candle.High, &candle.Low, &candle.Volume)
		dfCandle.Candles = append(dfCandle.Candles, candle)
	}
	err = rows.Err()
	if err != nil {
		return
	}
	return dfCandle, nil
}
