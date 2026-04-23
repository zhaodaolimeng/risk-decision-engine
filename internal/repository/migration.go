package repository

import (
	"risk-decision-engine/internal/repository/model"

	"gorm.io/gorm"
)

// AutoMigrate 自动迁移数据库
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.RuleConfig{},
		&model.DecisionRecord{},
		&model.RecordingSession{},
		&model.ReplaySession{},
	)
}
