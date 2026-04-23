// Risk Decision Engine API
//
// 风险决策引擎 - 沙盒服务
// 包含规则执行、流量记录、流量回放、规则比对等功能
//
// Schemes: http
// Host: localhost:8080
// BasePath: /
// Version: 1.0.0
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
//
// swagger:meta
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"risk-decision-engine/internal/api/errors"
	"risk-decision-engine/internal/api/handler"
	"risk-decision-engine/internal/api/middleware"
	"risk-decision-engine/internal/sandbox"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	recorder      *sandbox.Recorder
	replayer      *sandbox.Replayer
	comparator    *sandbox.RuleComparator
	configExecutor *sandbox.ConfigurableExecutor
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// DecisionRequest 决策请求
type DecisionRequest struct {
	RequestID  string                 `json:"requestId" binding:"required"`
	BusinessID string                 `json:"businessId"`
	Data       map[string]interface{} `json:"data" binding:"required"`
}

// DecisionResponse 决策响应
type DecisionResponse struct {
	DecisionID     string      `json:"decisionId"`
	BusinessID     string      `json:"businessId"`
	Decision       string      `json:"decision"`
	DecisionCode   string      `json:"decisionCode"`
	DecisionReason string      `json:"decisionReason"`
	RuleResults    interface{} `json:"ruleResults,omitempty"`
	DurationMs     int64       `json:"durationMs"`
}

// StartRecordingRequest 开始记录请求
type StartRecordingRequest struct {
	Name string `json:"name" binding:"required"`
}

// StartReplayOptionsRequest 开始回放请求（带选项）
type StartReplayOptionsRequest struct {
	Name           string     `json:"name" binding:"required"`
	RecordingID    string     `json:"recordingId" binding:"required"`
	RuleConfigPath string     `json:"ruleConfigPath"`
	TimeFrom       *time.Time `json:"timeFrom"`
	TimeTo         *time.Time `json:"timeTo"`
}

// CompareRulesRequest 规则比对请求
type CompareRulesRequest struct {
	Name          string `json:"name" binding:"required"`
	OldConfigPath string `json:"oldConfigPath" binding:"required"`
	NewConfigPath string `json:"newConfigPath" binding:"required"`
}

func initRules(configPath string) error {
	var err error
	configExecutor, err = sandbox.NewConfigurableExecutor(configPath)
	if err != nil {
		return err
	}

	rules := configExecutor.GetRules()
	fmt.Printf("✓ 规则加载成功，共 %d 条规则\n", len(rules))
	for _, r := range rules {
		fmt.Printf("  - [%s] %s: %s\n", r.ID, r.Name, r.Expression)
	}
	return nil
}

// HealthCheck godoc
// @Summary 健康检查
// @Description 检查服务是否正常运行
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func healthCheck(c *gin.Context) {
	middleware.RespondSuccess(c, HealthResponse{
		Status:  "ok",
		Service: "risk-decision-engine-sandbox",
		Version: "1.0.0",
	})
}

// ReloadRules godoc
// @Summary 重载规则配置
// @Description 重新加载规则配置文件
// @Tags 规则
// @Accept json
// @Produce json
// @Param config body object{configPath=string} true "配置路径"
// @Success 200 {object} object{code=string,message=string,data=object{ruleCount=int}}
// @Failure 400 {object} object{code=string,message=string}
// @Failure 500 {object} object{code=string,message=string}
// @Router /api/v1/rules/reload [post]
func reloadRules(c *gin.Context) {
	var req struct {
		ConfigPath string `json:"configPath"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithInvalidParams(c, err.Error())
		return
	}

	configPath := req.ConfigPath
	if configPath == "" {
		middleware.AbortWithError(c, errors.New(errors.CodeInvalidParams, "需要指定配置路径"))
		return
	}

	if err := configExecutor.LoadConfig(configPath); err != nil {
		middleware.AbortWithError(c, errors.Wrap(errors.CodeInternalError, err, "重载规则失败"))
		return
	}

	middleware.RespondSuccess(c, gin.H{
		"ruleCount": len(configExecutor.GetRules()),
	})
}

// ExecuteDecision godoc
// @Summary 执行决策
// @Description 根据输入数据执行规则决策
// @Tags 决策
// @Accept json
// @Produce json
// @Param request body DecisionRequest true "决策请求"
// @Success 200 {object} object{code=string,message=string,data=DecisionResponse}
// @Failure 400 {object} object{code=string,message=string}
// @Failure 500 {object} object{code=string,message=string}
// @Router /api/v1/decision/execute [post]
func executeDecision(c *gin.Context) {
	var req DecisionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithInvalidParams(c, err.Error())
		return
	}

	decisionID := "dec_" + uuid.NewString()[:8]
	start := time.Now()

	fmt.Printf("\n[决策请求] requestId=%s, decisionId=%s\n", req.RequestID, decisionID)

	// 构建事实数据
	fact := buildFactFromInput(req.Data)
	factJSON, _ := json.Marshal(fact)
	fmt.Printf("[事实数据] %s\n", string(factJSON))

	// 执行决策（使用可配置执行器）
	decision, decisionCode, decisionReason, ruleResults, modelResult, duration, err := configExecutor.Execute(
		req.RequestID,
		req.BusinessID,
		req.Data,
	)
	if err != nil {
		fmt.Printf("[决策错误] %v\n", err)
		middleware.AbortWithError(c, errors.Wrap(errors.CodeInternalError, err, "规则执行失败"))
		return
	}

	fmt.Printf("[决策结果] decisionId=%s, decision=%s, code=%s, reason=%s, duration=%v\n",
		decisionID, decision, decisionCode, decisionReason, duration)

	// 如果有活跃的记录会话，则记录
	if recorder != nil {
		if activeSession := recorder.GetActiveSession(); activeSession != nil {
			recorder.Record(
				req.RequestID,
				req.BusinessID,
				req.Data,
				decision,
				decisionCode,
				decisionReason,
				ruleResults,
				modelResult,
				duration,
			)
		}
	}

	middleware.RespondSuccess(c, DecisionResponse{
		DecisionID:     decisionID,
		BusinessID:     req.BusinessID,
		Decision:       decision,
		DecisionCode:   decisionCode,
		DecisionReason: decisionReason,
		RuleResults:    ruleResults,
		DurationMs:     duration.Milliseconds(),
	})
}

func buildFactFromInput(input map[string]interface{}) map[string]interface{} {
	fact := make(map[string]interface{})

	// 直接复制顶层字段
	for k, v := range input {
		fact[k] = v
	}

	// 从applicant中提取字段到顶层
	if applicant, ok := input["applicant"].(map[string]interface{}); ok {
		for k, v := range applicant {
			fact[k] = v
		}
	}

	// 处理嵌套字段
	flattenNestedFields(fact, input, "")

	return fact
}

func flattenNestedFields(fact, data map[string]interface{}, prefix string) {
	for k, v := range data {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		if nested, ok := v.(map[string]interface{}); ok {
			flattenNestedFields(fact, nested, key)
		} else {
			fact[key] = v
		}
	}
}

// ListRules godoc
// @Summary 获取规则列表
// @Description 获取当前加载的所有规则
// @Tags 规则
// @Accept json
// @Produce json
// @Success 200 {object} object{code=string,message=string,data=object{rules=[]object{id=string,name=string,expression=string}}}
// @Router /api/v1/rules [get]
func listRules(c *gin.Context) {
	type RuleInfo struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Expression string `json:"expression"`
	}

	rules := configExecutor.GetRules()
	var ruleList []RuleInfo
	for _, r := range rules {
		ruleList = append(ruleList, RuleInfo{
			ID:         r.ID,
			Name:       r.Name,
			Expression: r.Expression,
		})
	}

	middleware.RespondSuccess(c, gin.H{
		"rules": ruleList,
	})
}

func main() {
	fmt.Println("========================================")
	fmt.Println("   风险决策引擎 - 沙盒服务")
	fmt.Println("   支持流量记录、回放、规则比对")
	fmt.Println("========================================")
	fmt.Println()

	// 初始化沙盒组件
	recorder = sandbox.NewRecorder("sandbox/records")
	replayer = sandbox.NewReplayer(recorder)
	comparator = sandbox.NewRuleComparator()

	// 使用简单用例的规则配置
	configPath := "test/cases/simple/01-age-rule/config/rule.yaml"

	if err := initRules(configPath); err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("服务说明:")
	fmt.Println("  - 服务端口: 8080")
	fmt.Println("  - Swagger文档: http://localhost:8080/swagger/index.html")
	fmt.Println()
	fmt.Println("核心API:")
	fmt.Println("  - 健康检查: GET /health")
	fmt.Println("  - 规则列表: GET /api/v1/rules")
	fmt.Println("  - 执行决策: POST /api/v1/decision/execute")
	fmt.Println("  - 重载规则: POST /api/v1/rules/reload")
	fmt.Println()
	fmt.Println("沙盒API - 流量记录:")
	fmt.Println("  - 开始记录: POST /api/v1/sandbox/record/start")
	fmt.Println("  - 停止记录: POST /api/v1/sandbox/record/stop")
	fmt.Println("  - 会话列表: GET /api/v1/sandbox/record/sessions")
	fmt.Println()
	fmt.Println("沙盒API - 流量回放:")
	fmt.Println("  - 开始回放: POST /api/v1/sandbox/replay/start")
	fmt.Println("  - 开始回放(带选项): POST /api/v1/sandbox/replay/start/options")
	fmt.Println("  - 回放报告: GET /api/v1/sandbox/replay/sessions/{id}/report")
	fmt.Println()
	fmt.Println("按 Ctrl+C 停止服务")
	fmt.Println("========================================")
	fmt.Println()

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// 健康检查
	router.GET("/health", healthCheck)

	// Swagger文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 核心API
	router.GET("/api/v1/rules", listRules)
	router.POST("/api/v1/rules/reload", reloadRules)
	router.POST("/api/v1/decision/execute", executeDecision)

	// 沙盒API
	sandboxHandler := handler.NewSandboxHandler(recorder, replayer, comparator, configExecutor)
	sandboxHandler.RegisterRoutes(router)

	fmt.Println("服务启动: 0.0.0.0:8080")
	if err := router.Run("0.0.0.0:8080"); err != nil {
		fmt.Printf("服务启动失败: %v\n", err)
	}
}
