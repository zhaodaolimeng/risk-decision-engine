package repository

import (
	"risk-decision-engine/internal/repository/model"

	"gorm.io/gorm"
)

// RuleConfigRepository 规则配置仓储
type RuleConfigRepository struct {
	db *gorm.DB
}

// NewRuleConfigRepository 创建规则配置仓储
func NewRuleConfigRepository(db *gorm.DB) *RuleConfigRepository {
	return &RuleConfigRepository{db: db}
}

// Create 创建规则配置
func (r *RuleConfigRepository) Create(config *model.RuleConfig) error {
	return r.db.Create(config).Error
}

// Update 更新规则配置
func (r *RuleConfigRepository) Update(config *model.RuleConfig) error {
	return r.db.Save(config).Error
}

// FindByRuleID 根据规则ID查找
func (r *RuleConfigRepository) FindByRuleID(ruleID string) (*model.RuleConfig, error) {
	var config model.RuleConfig
	err := r.db.Where("rule_id = ?", ruleID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// FindByRuleIDAndVersion 根据规则ID和版本查找
func (r *RuleConfigRepository) FindByRuleIDAndVersion(ruleID, version string) (*model.RuleConfig, error) {
	var config model.RuleConfig
	err := r.db.Where("rule_id = ? AND version = ?", ruleID, version).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// ListActive 获取所有激活的规则配置
func (r *RuleConfigRepository) ListActive() ([]*model.RuleConfig, error) {
	var configs []*model.RuleConfig
	err := r.db.Where("is_active = ? AND status = ?", true, "ACTIVE").Order("priority DESC").Find(&configs).Error
	return configs, err
}

// ListAll 获取所有规则配置
func (r *RuleConfigRepository) ListAll() ([]*model.RuleConfig, error) {
	var configs []*model.RuleConfig
	err := r.db.Order("rule_id, version DESC").Find(&configs).Error
	return configs, err
}

// DeactivateByRuleID 停用某个规则的所有版本
func (r *RuleConfigRepository) DeactivateByRuleID(ruleID string) error {
	return r.db.Model(&model.RuleConfig{}).Where("rule_id = ?", ruleID).Update("is_active", false).Error
}
