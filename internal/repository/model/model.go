package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// RuleConfig 规则配置
type RuleConfig struct {
	BaseModel
	RuleID      string `gorm:"column:rule_id;uniqueIndex;size:64" json:"ruleId"`
	Version     string `gorm:"column:version;size:32" json:"version"`
	Name        string `gorm:"column:name;size:128" json:"name"`
	Description string `gorm:"column:description;size:512" json:"description"`
	Type        string `gorm:"column:type;size:32" json:"type"`
	Priority    int    `gorm:"column:priority" json:"priority"`
	Status      string `gorm:"column:status;size:32" json:"status"`
	Condition   string `gorm:"column:condition;type:text" json:"condition"`   // JSON格式
	Actions     string `gorm:"column:actions;type:text" json:"actions"`     // JSON格式
	ConfigPath  string `gorm:"column:config_path;size:256" json:"configPath"` // 来源配置文件路径
	IsActive    bool   `gorm:"column:is_active;default:true" json:"isActive"`
}

// TableName 表名
func (RuleConfig) TableName() string {
	return "rule_configs"
}

// DecisionRecord 决策记录
type DecisionRecord struct {
	BaseModel
	RequestID     string                 `gorm:"column:request_id;uniqueIndex;size:64" json:"requestId"`
	BusinessID    string                 `gorm:"column:business_id;index;size:64" json:"businessId"`
	Decision      string                 `gorm:"column:decision;size:32" json:"decision"`
	DecisionCode  string                 `gorm:"column:decision_code;size:64" json:"decisionCode"`
	DecisionReason string                `gorm:"column:decision_reason;size:512" json:"decisionReason"`
	InputData     string                 `gorm:"column:input_data;type:text" json:"inputData"`     // JSON格式
	RuleResults   string                 `gorm:"column:rule_results;type:text" json:"ruleResults"` // JSON格式
	ModelResults  string                 `gorm:"column:model_results;type:text" json:"modelResults,omitempty"`
	DurationMs    int64                  `gorm:"column:duration_ms" json:"durationMs"`
	RuleVersion    string                `gorm:"column:rule_version;size:64" json:"ruleVersion,omitempty"`
	SessionID      string                `gorm:"column:session_id;index;size:64" json:"sessionId,omitempty"`
}

// TableName 表名
func (DecisionRecord) TableName() string {
	return "decision_records"
}

// RecordingSession 记录会话
type RecordingSession struct {
	BaseModel
	SessionID   string     `gorm:"column:session_id;uniqueIndex;size:64" json:"sessionId"`
	Name        string     `gorm:"column:name;size:128" json:"name"`
	StartTime   time.Time  `gorm:"column:start_time" json:"startTime"`
	EndTime     *time.Time `gorm:"column:end_time" json:"endTime,omitempty"`
	IsActive    bool       `gorm:"column:is_active;default:false" json:"isActive"`
	RecordCount int        `gorm:"column:record_count;default:0" json:"recordCount"`
}

// TableName 表名
func (RecordingSession) TableName() string {
	return "recording_sessions"
}

// ReplaySession 回放会话
type ReplaySession struct {
	BaseModel
	SessionID   string     `gorm:"column:session_id;uniqueIndex;size:64" json:"sessionId"`
	Name        string     `gorm:"column:name;size:128" json:"name"`
	RecordingID string     `gorm:"column:recording_id;index;size:64" json:"recordingId"`
	ConfigPath  string     `gorm:"column:config_path;size:256" json:"configPath,omitempty"`
	StartTime   time.Time  `gorm:"column:start_time" json:"startTime"`
	EndTime     *time.Time `gorm:"column:end_time" json:"endTime,omitempty"`
	IsActive    bool       `gorm:"column:is_active;default:false" json:"isActive"`
	Report      string     `gorm:"column:report;type:text" json:"report,omitempty"` // JSON格式
}

// TableName 表名
func (ReplaySession) TableName() string {
	return "replay_sessions"
}
