package sandbox

import (
	"time"
)

// RecordedRequest 记录的请求
type RecordedRequest struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	RequestID   string                 `json:"requestId"`
	BusinessID  string                 `json:"businessId"`
	Data        map[string]interface{} `json:"data"`
}

// RecordedResponse 记录的响应
type RecordedResponse struct {
	ID             string      `json:"id"`
	RequestID      string      `json:"requestId"`
	Timestamp      time.Time   `json:"timestamp"`
	Decision       string      `json:"decision"`
	DecisionCode   string      `json:"decisionCode"`
	DecisionReason string      `json:"decisionReason"`
	RuleResults    interface{} `json:"ruleResults,omitempty"`
	ModelResult    interface{} `json:"modelResult,omitempty"`
	DurationMs     int64       `json:"durationMs"`
}

// TrafficRecord 完整流量记录
type TrafficRecord struct {
	Request  *RecordedRequest  `json:"request"`
	Response *RecordedResponse `json:"response"`
}

// ReplayResult 回放结果
type ReplayResult struct {
	RequestID        string            `json:"requestId"`
	OriginalDecision string            `json:"originalDecision"`
	NewDecision      string            `json:"newDecision"`
	DecisionMatch    bool              `json:"decisionMatch"`
	OriginalCode     string            `json:"originalCode"`
	NewCode          string            `json:"newCode"`
	CodeMatch        bool              `json:"codeMatch"`
	OriginalReason   string            `json:"originalReason"`
	NewReason        string            `json:"newReason"`
	DurationDiffMs   int64             `json:"durationDiffMs"`
}

// ReplayReport 回放报告
type ReplayReport struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	StartTime     time.Time      `json:"startTime"`
	EndTime       time.Time      `json:"endTime"`
	TotalRequests int            `json:"totalRequests"`
	MatchedCount  int            `json:"matchedCount"`
	MismatchedCount int           `json:"mismatchedCount"`
	MatchRate     float64        `json:"matchRate"`
	Results       []*ReplayResult `json:"results"`
	Mismatches    []*ReplayResult `json:"mismatches,omitempty"`
}

// RecordingSession 记录会话
type RecordingSession struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	StartTime   time.Time `json:"startTime"`
	EndTime     *time.Time `json:"endTime,omitempty"`
	IsActive    bool      `json:"isActive"`
	RecordCount int       `json:"recordCount"`
}

// ReplaySession 回放会话
type ReplaySession struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	RecordingID  string             `json:"recordingId"`
	ConfigPath   string             `json:"configPath,omitempty"`
	StartTime    time.Time          `json:"startTime"`
	EndTime      *time.Time         `json:"endTime,omitempty"`
	IsActive     bool               `json:"isActive"`
	Report       *ReplayReport      `json:"report,omitempty"`
	Options      *ReplayOptions     `json:"options,omitempty"`
}

// ReplayOptions 回放选项
type ReplayOptions struct {
	// 时段筛选
	TimeFrom *time.Time `json:"timeFrom,omitempty"`
	TimeTo   *time.Time `json:"timeTo,omitempty"`

	// 规则配置
	RuleConfigPath string `json:"ruleConfigPath,omitempty"`

	// 其他选项
	MaxParallel int  `json:"maxParallel,omitempty"`
	DryRun      bool `json:"dryRun,omitempty"`
}
