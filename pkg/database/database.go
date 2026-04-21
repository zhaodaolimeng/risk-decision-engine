package database

import (
	"fmt"
	"time"

	"risk-decision-engine/pkg/config"
	"risk-decision-engine/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	globalDB *gorm.DB
)

// Init 初始化数据库
func Init(cfg *config.MySQLConfig) error {
	// 创建 GORM 配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(cfg.DSN()), gormConfig)
	if err != nil {
		return fmt.Errorf("connect mysql: %w", err)
	}

	// 获取底层 SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime * time.Second)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}

	globalDB = db
	logger.Info("mysql connected")

	return nil
}

// DB 获取数据库实例
func DB() *gorm.DB {
	if globalDB == nil {
		panic("database not initialized")
	}
	return globalDB
}

// Close 关闭数据库连接
func Close() error {
	if globalDB != nil {
		sqlDB, err := globalDB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
