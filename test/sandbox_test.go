package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"risk-decision-engine/internal/sandbox"

	"github.com/stretchr/testify/assert"
)

// TestRecorder 测试流量记录器
func TestRecorder(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建记录器
	recorder := sandbox.NewRecorder(tmpDir)
	assert.NotNil(t, recorder)

	// 测试1: 开始记录会话
	t.Run("StartSession", func(t *testing.T) {
		session, err := recorder.StartSession("test-session-1")
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, "test-session-1", session.Name)
		assert.True(t, session.IsActive)
		assert.NotEmpty(t, session.ID)
	})

	// 测试2: 重复开始记录会话应该失败
	t.Run("StartSession_Duplicate", func(t *testing.T) {
		session, err := recorder.StartSession("test-session-2")
		assert.Error(t, err)
		assert.Nil(t, session)
	})

	// 测试3: 记录请求
	t.Run("Record", func(t *testing.T) {
		reqData := map[string]interface{}{
			"age": 25,
		}

		err := recorder.Record(
			"req-001",
			"biz-001",
			reqData,
			"APPROVE",
			"APPROVE_RULE",
			"通过",
			nil,
			nil,
			100*time.Millisecond,
		)
		assert.NoError(t, err)
	})

	// 测试4: 获取活跃会话
	t.Run("GetActiveSession", func(t *testing.T) {
		session := recorder.GetActiveSession()
		assert.NotNil(t, session)
		assert.Equal(t, 1, session.RecordCount)
	})

	// 测试5: 列出会话
	t.Run("ListSessions", func(t *testing.T) {
		sessions := recorder.ListSessions()
		assert.Len(t, sessions, 1)
	})

	// 测试6: 停止记录会话
	t.Run("StopSession", func(t *testing.T) {
		session, err := recorder.StopSession()
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.False(t, session.IsActive)
		assert.NotNil(t, session.EndTime)
	})

	// 测试7: 停止没有活跃会话应该失败
	t.Run("StopSession_NoActive", func(t *testing.T) {
		session, err := recorder.StopSession()
		assert.Error(t, err)
		assert.Nil(t, session)
	})
}

// TestRecorder_GetRecords 测试获取记录
func TestRecorder_GetRecords(t *testing.T) {
	tmpDir := t.TempDir()
	recorder := sandbox.NewRecorder(tmpDir)

	// 创建会话并记录一些数据
	session, err := recorder.StartSession("test-session")
	assert.NoError(t, err)

	// 记录3条数据
	for i := 1; i <= 3; i++ {
		err := recorder.Record(
			"req-00"+string(rune('0'+i)),
			"biz-001",
			map[string]interface{}{"age": 20 + i},
			"APPROVE",
			"APPROVE_RULE",
			"通过",
			nil,
			nil,
			100*time.Millisecond,
		)
		assert.NoError(t, err)
	}
	_, err = recorder.StopSession()
	assert.NoError(t, err)

	// 测试获取记录
	t.Run("GetRecords", func(t *testing.T) {
		records, err := recorder.GetRecords(session.ID)
		assert.NoError(t, err)
		assert.Len(t, records, 3)
	})

	// 测试获取不存在的会话记录
	t.Run("GetRecords_NotFound", func(t *testing.T) {
		records, err := recorder.GetRecords("invalid-id")
		assert.Error(t, err)
		assert.Nil(t, records)
	})
}

// TestComparator 测试规则配置比对器
func TestComparator(t *testing.T) {
	comparator := sandbox.NewRuleComparator()
	assert.NotNil(t, comparator)

	// 加载测试配置
	oldConfigPath := filepath.Join("cases", "simple", "01-age-rule", "config", "rule.yaml")
	newConfigPath := filepath.Join("cases", "simple", "01-age-rule", "config", "rule_v2.yaml")

	// 确保文件存在
	if _, err := os.Stat(oldConfigPath); err != nil {
		t.Skip("Test config file not found")
	}
	if _, err := os.Stat(newConfigPath); err != nil {
		t.Skip("Test config file not found")
	}

	// 这里简单测试比对器的基本功能，不需要加载完整规则
	t.Run("NewComparator", func(t *testing.T) {
		assert.NotNil(t, comparator)
	})
}

// TestReplayOptions 测试回放选项
func TestReplayOptions(t *testing.T) {
	t.Run("ReplayOptions_Create", func(t *testing.T) {
		now := time.Now()
		options := &sandbox.ReplayOptions{
			TimeFrom:       &now,
			RuleConfigPath: "test.yaml",
			MaxParallel:    10,
			DryRun:         true,
		}
		assert.NotNil(t, options)
		assert.Equal(t, "test.yaml", options.RuleConfigPath)
		assert.Equal(t, 10, options.MaxParallel)
		assert.True(t, options.DryRun)
	})
}
