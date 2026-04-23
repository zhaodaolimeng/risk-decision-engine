package repository

import (
	"risk-decision-engine/internal/repository/model"

	"gorm.io/gorm"
)

// DecisionRecordRepository 决策记录仓储
type DecisionRecordRepository struct {
	db *gorm.DB
}

// NewDecisionRecordRepository 创建决策记录仓储
func NewDecisionRecordRepository(db *gorm.DB) *DecisionRecordRepository {
	return &DecisionRecordRepository{db: db}
}

// Create 创建决策记录
func (r *DecisionRecordRepository) Create(record *model.DecisionRecord) error {
	return r.db.Create(record).Error
}

// FindByRequestID 根据请求ID查找
func (r *DecisionRecordRepository) FindByRequestID(requestID string) (*model.DecisionRecord, error) {
	var record model.DecisionRecord
	err := r.db.Where("request_id = ?", requestID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// FindByBusinessID 根据业务ID查找
func (r *DecisionRecordRepository) FindByBusinessID(businessID string, limit int) ([]*model.DecisionRecord, error) {
	var records []*model.DecisionRecord
	err := r.db.Where("business_id = ?", businessID).Order("created_at DESC").Limit(limit).Find(&records).Error
	return records, err
}

// FindBySessionID 根据会话ID查找
func (r *DecisionRecordRepository) FindBySessionID(sessionID string) ([]*model.DecisionRecord, error) {
	var records []*model.DecisionRecord
	err := r.db.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&records).Error
	return records, err
}

// ListRecent 列出最近的记录
func (r *DecisionRecordRepository) ListRecent(limit int) ([]*model.DecisionRecord, error) {
	var records []*model.DecisionRecord
	err := r.db.Order("created_at DESC").Limit(limit).Find(&records).Error
	return records, err
}

// CountByDecision 按决策结果统计
func (r *DecisionRecordRepository) CountByDecision() (map[string]int64, error) {
	var results []struct {
		Decision string
		Count    int64
	}
	err := r.db.Model(&model.DecisionRecord{}).Select("decision, count(*) as count").Group("decision").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	countMap := make(map[string]int64)
	for _, r := range results {
		countMap[r.Decision] = r.Count
	}
	return countMap, nil
}
