package model

import (
	"database/sql"
	"fmt"
	logstd "log"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/v03413/bepusdt/app/conf"
	"gorm.io/gorm"

	"github.com/v03413/bepusdt/app/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB
var err error

func Init() error {
	if len(conf.GetConfig().MySQL.DSN) > 0 {
		if err := initMysql(); err != nil {
			return err
		}
	} else {
		var path = conf.GetSqlitePath()
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {

			return fmt.Errorf("创建数据库目录失败：%w", err)
		}

		DB, err = gorm.Open(sqlite.Open(path), gormConfig())
		if err != nil {

			return fmt.Errorf("数据库初始化失败：%w", err)
		}
		if conf.GetDebug() {
			DB = DB.Debug()
		}
	}

	if err = AutoMigrate(); err != nil {

		return fmt.Errorf("数据库结构迁移失败：%w", err)
	}

	addStartWalletAddress()

	return nil
}

func AutoMigrate() error {

	return DB.AutoMigrate(&WalletAddress{}, &TradeOrders{}, &NotifyRecord{}, &Config{}, &Webhook{})
}

func gormConfig() *gorm.Config {
	return &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   conf.GetConfig().MySQL.TablePrefix,
			SingularTable: true,
		},
		Logger: logger.New(logstd.New(os.Stdout, "\r\n", logstd.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		}),
	}
}

func initMysql() error {
	var err error
	cfg := conf.GetConfig().MySQL
	DB, err = gorm.Open(mysql.Open(cfg.DSN), gormConfig())
	if err != nil {
		return err
	}
	if conf.GetDebug() {
		DB = DB.Debug()
	}
	var sqlDB *sql.DB
	sqlDB, err = DB.DB()
	if err != nil {
		return fmt.Errorf("mysql get DB, err=%v", err)
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour * time.Duration(cfg.MaxLifeTime))
	err = sqlDB.Ping()
	if err != nil {
		return fmt.Errorf("mysql connDB err: %v", err)
	}
	log.Info("mysql connDB success")
	return err
}
