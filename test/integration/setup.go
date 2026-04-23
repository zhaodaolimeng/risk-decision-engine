package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"risk-decision-engine/internal/api/middleware"
	"risk-decision-engine/internal/api/handler"
	"risk-decision-engine/internal/engine/rule"
	"risk-decision-engine/internal/sandbox"

	"github.com/gin-gonic/gin"
)

// TestSetup 测试环境设置
type TestSetup struct {
	Router         *gin.Engine
	Recorder       *sandbox.Recorder
	Replayer       *sandbox.Replayer
	Comparator     *sandbox.RuleComparator
	Executor       *sandbox.ConfigurableExecutor
	Ctx            context.Context
}

// SetupTest 初始化测试环境
func SetupTest(t *testing.T) *TestSetup {
	gin.SetMode(gin.TestMode)

	// 初始化沙盒组件
	recorder := sandbox.NewRecorder("") // 使用内存模式
	replayer := sandbox.NewReplayer(recorder)
	comparator := sandbox.NewRuleComparator()

	// 加载测试规则
	testConfigPath := "../cases/simple/01-age-rule/config/rule.yaml"
	executor, err := sandbox.NewConfigurableExecutor(testConfigPath)
	if err != nil {
		t.Logf("Warning: Failed to load test config, using empty: %v", err)
	}

	// 创建路由
	router := gin.Default()

	// 注册健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 注册沙盒API
	sandboxHandler := handler.NewSandboxHandler(recorder, replayer, comparator, executor)
	sandboxHandler.RegisterRoutes(router)

	// 注册决策API
	router.POST("/api/v1/decision/execute", func(c *gin.Context) {
		var req struct {
			RequestID string                 `json:"requestId"`
			BusinessID string                `json:"businessId"`
			Data      map[string]interface{} `json:"data"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithInvalidParams(c, err.Error())
			return
		}

		decision, code, reason, ruleResults, modelResult, duration, err := executor.Execute(
			req.RequestID,
			req.BusinessID,
			req.Data,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 记录到沙盒
		if activeSession := recorder.GetActiveSession(); activeSession != nil {
			recorder.Record(req.RequestID, req.BusinessID, req.Data, decision, code, reason, ruleResults, modelResult, duration)
		}

		middleware.RespondSuccess(c, gin.H{
			"decisionId": "dec-test",
			"decision": decision,
			"decisionCode": code,
			"decisionReason": reason,
			"ruleResults": ruleResults,
			"durationMs": duration.Milliseconds(),
		})
	})

	return &TestSetup{
		Router:     router,
		Recorder:   recorder,
		Replayer:   replayer,
		Comparator: comparator,
		Executor:   executor,
		Ctx:        context.Background(),
	}
}

// MakeRequest 发起测试请求
func (ts *TestSetup) MakeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			panic(err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	return w
}

// ParseResponse 解析响应
func ParseResponse[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	var result struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    T      `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	return result.Data
}

// LoadTestRules 加载测试规则
func LoadTestRules(t *testing.T) ([]*rule.SimpleRule, error) {
	testConfigPath := "../cases/simple/01-age-rule/config/rule.yaml"
	return rule.LoadRulesFromFile(testConfigPath)
}
