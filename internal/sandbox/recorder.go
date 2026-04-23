package sandbox

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Recorder 流量记录器
type Recorder struct {
	sessions     map[string]*RecordingSession
	records      map[string][]*TrafficRecord
	mu           sync.RWMutex
	storageDir   string
	activeSession *RecordingSession
}

// NewRecorder 创建记录器
func NewRecorder(storageDir string) *Recorder {
	if storageDir == "" {
		storageDir = "./sandbox/records"
	}
	os.MkdirAll(storageDir, 0755)

	return &Recorder{
		sessions:   make(map[string]*RecordingSession),
		records:    make(map[string][]*TrafficRecord),
		storageDir: storageDir,
	}
}

// StartSession 开始记录会话
func (r *Recorder) StartSession(name string) (*RecordingSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.activeSession != nil && r.activeSession.IsActive {
		return nil, fmt.Errorf("already has active recording session: %s", r.activeSession.ID)
	}

	session := &RecordingSession{
		ID:        "rec_" + uuid.NewString()[:8],
		Name:      name,
		StartTime: time.Now(),
		IsActive:  true,
	}

	r.sessions[session.ID] = session
	r.records[session.ID] = make([]*TrafficRecord, 0)
	r.activeSession = session

	fmt.Printf("[Recorder] 会话已开始: %s (%s)\n", session.Name, session.ID)
	return session, nil
}

// StopSession 停止记录会话
func (r *Recorder) StopSession() (*RecordingSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.activeSession == nil || !r.activeSession.IsActive {
		return nil, fmt.Errorf("no active recording session")
	}

	session := r.activeSession
	now := time.Now()
	session.EndTime = &now
	session.IsActive = false
	session.RecordCount = len(r.records[session.ID])

	r.activeSession = nil

	// 保存到文件
	if err := r.saveSessionToFile(session); err != nil {
		fmt.Printf("[Recorder] 保存会话失败: %v\n", err)
	}

	fmt.Printf("[Recorder] 会话已停止: %s, 记录数: %d\n", session.Name, session.RecordCount)
	return session, nil
}

// Record 记录请求和响应
func (r *Recorder) Record(
	requestID, businessID string,
	reqData map[string]interface{},
	decision, decisionCode, decisionReason string,
	ruleResults, modelResult interface{},
	duration time.Duration,
) error {
	r.mu.RLock()
	session := r.activeSession
	r.mu.RUnlock()

	if session == nil || !session.IsActive {
		return nil // 没有活跃会话，不记录
	}

	recordID := uuid.NewString()

	req := &RecordedRequest{
		ID:         recordID,
		Timestamp:  time.Now(),
		RequestID:  requestID,
		BusinessID: businessID,
		Data:       reqData,
	}

	resp := &RecordedResponse{
		ID:             recordID,
		RequestID:      requestID,
		Timestamp:      time.Now(),
		Decision:       decision,
		DecisionCode:   decisionCode,
		DecisionReason: decisionReason,
		RuleResults:    ruleResults,
		ModelResult:    modelResult,
		DurationMs:     duration.Milliseconds(),
	}

	record := &TrafficRecord{
		Request:  req,
		Response: resp,
	}

	r.mu.Lock()
	r.records[session.ID] = append(r.records[session.ID], record)
	session.RecordCount = len(r.records[session.ID])
	r.mu.Unlock()

	return nil
}

// GetSession 获取会话
func (r *Recorder) GetSession(sessionID string) (*RecordingSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return session, nil
}

// GetRecords 获取会话的记录
func (r *Recorder) GetRecords(sessionID string) ([]*TrafficRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, ok := r.records[sessionID]
	if !ok {
		// 尝试从文件加载
		return r.loadRecordsFromFile(sessionID)
	}
	return records, nil
}

// GetRecordsByTimeRange 获取指定时间段的记录
func (r *Recorder) GetRecordsByTimeRange(
	sessionID string,
	timeFrom, timeTo *time.Time,
) ([]*TrafficRecord, error) {
	records, err := r.GetRecords(sessionID)
	if err != nil {
		return nil, err
	}

	if timeFrom == nil && timeTo == nil {
		return records, nil
	}

	var filtered []*TrafficRecord
	for _, record := range records {
		reqTime := record.Request.Timestamp

		// 检查起始时间
		if timeFrom != nil && reqTime.Before(*timeFrom) {
			continue
		}

		// 检查结束时间
		if timeTo != nil && reqTime.After(*timeTo) {
			continue
		}

		filtered = append(filtered, record)
	}

	return filtered, nil
}

// ListSessions 列出所有会话
func (r *Recorder) ListSessions() []*RecordingSession {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]*RecordingSession, 0, len(r.sessions))
	for _, s := range r.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// GetActiveSession 获取活跃会话
func (r *Recorder) GetActiveSession() *RecordingSession {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.activeSession
}

// 保存会话到文件
func (r *Recorder) saveSessionToFile(session *RecordingSession) error {
	records := r.records[session.ID]

	data := struct {
		Session *RecordingSession `json:"session"`
		Records []*TrafficRecord  `json:"records"`
	}{
		Session: session,
		Records: records,
	}

	filename := filepath.Join(r.storageDir, session.ID+".json")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// 从文件加载记录
func (r *Recorder) loadRecordsFromFile(sessionID string) ([]*TrafficRecord, error) {
	filename := filepath.Join(r.storageDir, sessionID+".json")
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data struct {
		Session *RecordingSession `json:"session"`
		Records []*TrafficRecord  `json:"records"`
	}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.sessions[sessionID] = data.Session
	r.records[sessionID] = data.Records
	r.mu.Unlock()

	return data.Records, nil
}
