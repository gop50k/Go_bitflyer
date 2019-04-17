package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/gemcook/gop50k-training/backend-2-go-fintech/Section20/config"
	// sqlite3 is SQLDriver アンスコで始まるimportはコメントを書かなければならない
	_ "github.com/mattn/go-sqlite3"
)

// base.go DB関連のfunctionなどを作成するファイル

// テーブルネームの指定
const (
	tableNameSignalEvents = "signal_events"
)

var DbConnection *sql.DB

// GetCandleTableName テーブルネームをGetするfunction
func GetCandleTableName(productCode string, duration time.Duration) string {
	return fmt.Sprintf("%s_%s", productCode, duration)
}

// DBスキーマの作成
func init() {
	var err error
	DbConnection, err = sql.Open(config.Config.SQLDriver, config.Config.DbName)
	if err != nil {
		log.Fatalln(err)
	}
	// クエリの作成
	// ビットコインの売買のイベントを書き込むテーブルを作成
	cmd := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		time DATETIME PRIMARY KEY NOT NULL,
		product_code STRING,
		side STRING,
		price FLOAT,
		size FLOAT)`, tableNameSignalEvents)

	// SQL実行
	DbConnection.Exec(cmd)

	for _, duration := range config.Config.Durations {
		tableName := GetCandleTableName(config.Config.ProductCode, duration)

		// キャンドルスティックの情報を入れるテーブルを作成
		c := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		time DATETIME PRIMARY KEY NOT NULL,
		open FLOAT,
		close FLOAT,
		high FLOAT,
		low FLOAT,
		volume FLOAT)`, tableName)
		DbConnection.Exec(c)
	}
}
