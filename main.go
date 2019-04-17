package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/app/controllers"
	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/app/models"
	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/config"
	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/utils"
)

func main() {
	// パフォーマンスが出るインディケーターのBest３を表示する
	df, _ := models.GetAllCandle(config.Config.ProductCode, time.Minute, 365)
	fmt.Printf("%+v\n", df.OptimizeParams())

	utils.LoggingSettings(config.Config.LogFile)

	// ストリーミングされたデータを表示
	controllers.StreamIngestionData()

	// キャンドルスティックチャートを表示
	log.Println(controllers.StartWebServer())
}
