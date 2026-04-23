package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSandboxWorkflow 测试完整的沙盒工作流
func TestSandboxWorkflow(t *testing.T) {
	ts := SetupTest(t)

	t.Run("1. 开始记录会话", func(t *testing.T) {
		req := map[string]string{"name": "test-session"}
		w := ts.MakeRequest("POST", "/api/v1/sandbox/record/start", req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("2. 执行多个决策请求（记录中）", func(t *testing.T) {
		testCases := []struct {
			requestID string
			age       int
		}{
			{"test-001", 25},
			{"test-002", 18},
			{"test-003", 30},
		}

		for _, tc := range testCases {
			req := map[string]interface{}{
				"requestId": tc.requestID,
				"businessId": "biz-001",
				"data": map[string]interface{}{
					"age": tc.age,
				},
			}
			w := ts.MakeRequest("POST", "/api/v1/decision/execute", req)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("3. 停止记录会话", func(t *testing.T) {
		w := ts.MakeRequest("POST", "/api/v1/sandbox/record/stop", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("4. 获取记录会话列表", func(t *testing.T) {
		w := ts.MakeRequest("GET", "/api/v1/sandbox/record/sessions", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("5. 开始回放", func(t *testing.T) {
		sessions := ts.Recorder.ListSessions()
		assert.GreaterOrEqual(t, len(sessions), 1)

		req := map[string]interface{}{
			"name": "replay-test",
			"recordingId": sessions[0].ID,
			"configPath": "../cases/simple/01-age-rule/config/rule_v2.yaml",
		}
		w := ts.MakeRequest("POST", "/api/v1/sandbox/replay/start/options", req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("6. 等待回放完成", func(t *testing.T) {
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("7. 获取回放会话列表", func(t *testing.T) {
		w := ts.MakeRequest("GET", "/api/v1/sandbox/replay/sessions", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestSandboxAPIs 测试沙盒API
func TestSandboxAPIs(t *testing.T) {
	ts := SetupTest(t)

	t.Run("健康检查", func(t *testing.T) {
		w := ts.MakeRequest("GET", "/health", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("规则比对", func(t *testing.T) {
		req := map[string]string{
			"name": "compare-test",
			"oldConfigPath": "../cases/simple/01-age-rule/config/rule.yaml",
			"newConfigPath": "../cases/simple/01-age-rule/config/rule_v2.yaml",
		}
		w := ts.MakeRequest("POST", "/api/v1/sandbox/diff/rules", req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestDecisionWorkflow 测试决策工作流
func TestDecisionWorkflow(t *testing.T) {
	ts := SetupTest(t)

	t.Run("通过年龄", func(t *testing.T) {
		req := map[string]interface{}{
			"requestId": "age-pass-001",
			"data": map[string]interface{}{
				"age": 25,
			},
		}
		w := ts.MakeRequest("POST", "/api/v1/decision/execute", req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("拒绝年龄（太小）", func(t *testing.T) {
		req := map[string]interface{}{
			"requestId": "age-reject-001",
			"data": map[string]interface{}{
				"age": 18,
			},
		}
		w := ts.MakeRequest("POST", "/api/v1/decision/execute", req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("拒绝年龄（太大）", func(t *testing.T) {
		req := map[string]interface{}{
			"requestId": "age-reject-002",
			"data": map[string]interface{}{
				"age": 65,
			},
		}
		w := ts.MakeRequest("POST", "/api/v1/decision/execute", req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
