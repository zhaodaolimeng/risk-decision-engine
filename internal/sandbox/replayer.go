package sandbox

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// DecisionExecutor 决策执行器接口
type DecisionExecutor interface {
	Execute(requestID, businessID string, data map[string]interface{}) (
		decision, decisionCode, decisionReason string,
		ruleResults, modelResult interface{},
		duration time.Duration,
		err error,
	)
}

// Replayer 流量回放器
type Replayer struct {
	recorder   *Recorder
	sessions   map[string]*ReplaySession
	mu         sync.RWMutex
}

// NewReplayer 创建回放器
func NewReplayer(recorder *Recorder) *Replayer {
	return &Replayer{
		recorder: recorder,
		sessions: make(map[string]*ReplaySession),
	}
}

// StartReplay 开始回放
func (r *Replayer) StartReplay(
	name, recordingID string,
	configPath string,
	executor DecisionExecutor,
) (*ReplaySession, error) {
	return r.StartReplayWithOptions(
		name,
		recordingID,
		&ReplayOptions{
			RuleConfigPath: configPath,
		},
		executor,
	)
}

// StartReplayWithOptions 开始回放（带选项）
func (r *Replayer) StartReplayWithOptions(
	name, recordingID string,
	options *ReplayOptions,
	executor DecisionExecutor,
) (*ReplaySession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if options == nil {
		options = &ReplayOptions{}
	}

	// 检查录制是否存在
	var records []*TrafficRecord
	var err error

	if options.TimeFrom != nil || options.TimeTo != nil {
		records, err = r.recorder.GetRecordsByTimeRange(recordingID, options.TimeFrom, options.TimeTo)
	} else {
		records, err = r.recorder.GetRecords(recordingID)
	}

	if err != nil {
		return nil, fmt.Errorf("recording not found: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no records in recording")
	}

	// 如果有配置路径，并且executor是可配置的，尝试加载配置
	if options.RuleConfigPath != "" {
		if cfgExecutor, ok := executor.(*ConfigurableExecutor); ok {
			if err := cfgExecutor.LoadConfig(options.RuleConfigPath); err != nil {
				fmt.Printf("[Replayer] 警告: 加载配置失败 %s: %v\n", options.RuleConfigPath, err)
			} else {
				fmt.Printf("[Replayer] 已加载配置: %s\n", options.RuleConfigPath)
			}
		}
	}

	session := &ReplaySession{
		ID:          "rep_" + uuid.NewString()[:8],
		Name:        name,
		RecordingID: recordingID,
		ConfigPath:  options.RuleConfigPath,
		Options:     options,
		StartTime:   time.Now(),
		IsActive:    true,
	}

	r.sessions[session.ID] = session

	// 打印回放信息
	fmt.Printf("[Replayer] 回放已开始: %s (%s)\n", session.Name, session.ID)
	fmt.Printf("[Replayer] 记录数: %d\n", len(records))
	if options.TimeFrom != nil {
		fmt.Printf("[Replayer] 从: %s\n", options.TimeFrom.Format(time.RFC3339))
	}
	if options.TimeTo != nil {
		fmt.Printf("[Replayer] 到: %s\n", options.TimeTo.Format(time.RFC3339))
	}
	if options.RuleConfigPath != "" {
		fmt.Printf("[Replayer] 规则配置: %s\n", options.RuleConfigPath)
	}

	// 异步执行回放
	go r.executeReplay(session, records, executor)

	return session, nil
}

// executeReplay 执行回放
func (r *Replayer) executeReplay(
	session *ReplaySession,
	records []*TrafficRecord,
	executor DecisionExecutor,
) {
	report := &ReplayReport{
		ID:            "rpt_" + uuid.NewString()[:8],
		Name:          session.Name,
		StartTime:     time.Now(),
		TotalRequests: len(records),
		Results:       make([]*ReplayResult, 0, len(records)),
		Mismatches:    make([]*ReplayResult, 0),
	}

	for _, record := range records {
		req := record.Request
		origResp := record.Response

		// 执行决策
		decision, decisionCode, decisionReason, ruleResults, modelResult, duration, err := executor.Execute(
			req.RequestID,
			req.BusinessID,
			req.Data,
		)

		if err != nil {
			fmt.Printf("[Replayer] 执行失败 [%s]: %v\n", req.RequestID, err)
			continue
		}

		// 比对结果
		result := &ReplayResult{
			RequestID:        req.RequestID,
			OriginalDecision: origResp.Decision,
			NewDecision:      decision,
			DecisionMatch:    origResp.Decision == decision,
			OriginalCode:     origResp.DecisionCode,
			NewCode:          decisionCode,
			CodeMatch:        origResp.DecisionCode == decisionCode,
			OriginalReason:   origResp.DecisionReason,
			NewReason:        decisionReason,
			DurationDiffMs:   duration.Milliseconds() - origResp.DurationMs,
		}

		report.Results = append(report.Results, result)

		if !result.DecisionMatch || !result.CodeMatch {
			report.Mismatches = append(report.Mismatches, result)
			fmt.Printf("[Replayer] 不匹配 [%s]: orig=%s/%s, new=%s/%s\n",
				req.RequestID, origResp.Decision, origResp.DecisionCode,
				decision, decisionCode)
		}
	}

	// 完成报告
	report.EndTime = time.Now()
	report.MatchedCount = report.TotalRequests - len(report.Mismatches)
	report.MismatchedCount = len(report.Mismatches)
	if report.TotalRequests > 0 {
		report.MatchRate = float64(report.MatchedCount) / float64(report.TotalRequests) * 100
	}

	r.mu.Lock()
	session.IsActive = false
	session.EndTime = &report.EndTime
	session.Report = report
	r.mu.Unlock()

	fmt.Printf("[Replayer] 回放完成: 总计=%d, 匹配=%d, 不匹配=%d, 匹配率=%.2f%%\n",
		report.TotalRequests, report.MatchedCount, report.MismatchedCount, report.MatchRate)
}

// GetSession 获取回放会话
func (r *Replayer) GetSession(sessionID string) (*ReplaySession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return session, nil
}

// ListSessions 列出所有回放会话
func (r *Replayer) ListSessions() []*ReplaySession {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]*ReplaySession, 0, len(r.sessions))
	for _, s := range r.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// GetReport 获取回放报告
func (r *Replayer) GetReport(sessionID string) (*ReplayReport, error) {
	session, err := r.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	if session.Report == nil {
		return nil, fmt.Errorf("report not ready")
	}
	return session.Report, nil
}
