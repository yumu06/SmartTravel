package TravelDate

import (
	"fmt"
	"net/url"
	"os"
	"time"
	"travel/TravelModel"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// @title InitDB
// @description	初始化数据库
// @auth	Snactop	2023-11-27	20:07
// @param	void	没有传入值
// @return	void	没有返回值
type SQLPool interface {
	SetMaxOpenConns(int)
	SetMaxIdleConns(int)
	SetConnMaxLifetime(time.Duration)
	SetConnMaxIdleTime(time.Duration)
}

type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func ConfigurePool(pool SQLPool, cfg PoolConfig) {
	pool.SetMaxOpenConns(cfg.MaxOpenConns)
	pool.SetMaxIdleConns(cfg.MaxIdleConns)
	pool.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	pool.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
}

func InitDB() error {
	host := viper.GetString("mysql.host")
	port := viper.GetString("mysql.port")
	database := viper.GetString("mysql.database")
	root := viper.GetString("mysql.root")
	password := viper.GetString("mysql.password")
	loc := viper.GetString("mysql.loc")

	args := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=%s",
		root,
		password,
		host,
		port,
		database,
		url.QueryEscape(loc))

	slowThreshold := viper.GetDuration("mysql.slow_threshold")
	if slowThreshold <= 0 {
		slowThreshold = 200 * time.Millisecond
	}
	gormLogger := logger.New(
		logWriter{writer: os.Stdout},
		logger.Config{SlowThreshold: slowThreshold, LogLevel: logger.Warn, Colorful: false},
	)
	db, err := gorm.Open(mysql.Open(args), &gorm.Config{Logger: gormLogger})
	if err != nil {
		return fmt.Errorf("connect database at %s: %w", host, err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql database: %w", err)
	}
	ConfigurePool(sqlDB, PoolConfig{
		MaxOpenConns:    viper.GetInt("mysql.max_open_conns"),
		MaxIdleConns:    viper.GetInt("mysql.max_idle_conns"),
		ConnMaxLifetime: viper.GetDuration("mysql.conn_max_lifetime"),
		ConnMaxIdleTime: viper.GetDuration("mysql.conn_max_idle_time"),
	})
	if err := db.AutoMigrate(&TravelModel.TraUser{}, &TravelModel.TraUserFoot{}, &TravelModel.TraUserFootStart{}, TravelModel.TraUserPostStart{}, &TravelModel.TraUserPostLike{}); err != nil {
		return fmt.Errorf("migrate user tables: %w", err)
	}

	if err := db.AutoMigrate(&TravelModel.TraFoot{}); err != nil {
		return fmt.Errorf("migrate foot tables: %w", err)
	}

	if err := db.AutoMigrate(&TravelModel.Post{}, &TravelModel.PostComment{}, &TravelModel.Notice{}, &TravelModel.ChatMessage{}); err != nil {
		return fmt.Errorf("migrate content tables: %w", err)
	}

	DB = db
	return nil
}

type logWriter struct {
	writer *os.File
}

func (w logWriter) Printf(format string, args ...interface{}) {
	fmt.Fprintf(w.writer, format+"\n", args...)
}

func GetDB() *gorm.DB {
	return DB
}
