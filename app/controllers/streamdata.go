package controllers

import (
	"log"

	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/app/models"
	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/bitflyer"
	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/config"
)

// StreamIngestionData データをストリーミングするfunction
func StreamIngestionData() {
	c := config.Config
	ai := NewAI(c.ProductCode, c.TradeDuration, c.DataLimit, c.UsePercent, c.StopLimitPercent, c.BackTest)

	var tickerChannel = make(chan bitflyer.Ticker)
	apiClient := bitflyer.New(config.Config.APIKey, config.Config.APISecret)
	go apiClient.GetRealTimeTicker(config.Config.ProductCode, tickerChannel)

	go func() {
		for ticker := range tickerChannel {
			// 秒、分、時間ごとにデータの書き込みを行う
			log.Printf("action=StreamIngestionData, %v", ticker)
			for _, duration := range config.Config.Durations {
				isCreated := models.CreateCandleWithDuration(ticker, ticker.ProductCode, duration)
				if isCreated == true && duration == config.Config.TradeDuration {
					ai.Trade()
				}
			}
		}
	}()
}
