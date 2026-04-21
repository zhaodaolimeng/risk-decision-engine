package main

import (
	"fmt"
	"net/http"
	"sync"

	"risk-decision-engine/internal/engine/rule"
	"risk-decision-engine/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	ruleService *RuleService
	once        sync.Once
)

// RuleService 规则服务
type RuleService struct {
	rules map[string]*rule.Rule
	mu    sync.RWMutex
}

func NewRuleService() *RuleService {
	return &RuleService{
		rules: make(map[string]*rule.Rule),
	}
}

func (s *RuleService) LoadAgeRule() error {
	ruleJSON := []byte(`{
		"ruleId": "R001",
		"version": "1.0",
		"name": "年龄准入规则",
		"description": "申请人年龄必须在21-60岁之间",
		"type": "BOOLEAN",
		"priority": 100,
		"status": "ACTIVE",
		"condition": {
			"operator": "AND",
			"expressions": [
				{
					"field": "age",
					"operator": ">=",
					"value": 21
				},
				{
					"field": "age",
					"operator": "<=",
					"value": 60
				}
			]
		},
		"actions": {
			"true": {
				"result": "PASS"
			},
			"false": {
				"result": "REJECT",
				"reason": "年龄不符合要求，需在21-60岁之间"
			}
		}
	}`)

	r, err := rule.LoadRuleFromJSON(ruleJSON)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.rules[r.RuleID] = r
	s.mu.Unlock()

	fmt.Printf("✓ 规则加载成功: %s\n", r.Name)
	return nil
}

func (s *RuleService) GetRule(id string) (*rule.Rule, bool) {
	s.mu.RLock()
	r, ok := s.rules[id]
	s.mu.RUnlock()
	return r, ok
}

func initService() {
	once.Do(func() {
		ruleService = NewRuleService()
		if err := ruleService.LoadAgeRule(); err != nil {
			fmt.Printf("加载规则失败: %v\n", err)
		}
	})
}

func main() {
	fmt.Println("========================================")
	fmt.Println("   风险决策引擎服务 (完整版)")
	fmt.Println("========================================")
	fmt.Println()

	// 初始化服务
	initService()

	fmt.Println()
	fmt.Println("本机访问地址:")
	fmt.Println("  - 健康检查: http://192.168.1.27:8080/health")
	fmt.Println("  - 决策接口: http://192.168.1.27:8080/api/v1/decision/execute")
	fmt.Println()
	fmt.Println("测试用例:")
	fmt.Println("  - 年龄 20 岁: 应返回 REJECT")
	fmt.Println("  - 年龄 25 岁: 应返回 APPROVE")
	fmt.Println("  - 年龄 60 岁: 应返回 APPROVE")
	fmt.Println("  - 年龄 61 岁: 应返回 REJECT")
	fmt.Println()
	fmt.Println("服务监听: 0.0.0.0:8080")
	fmt.Println("按 Ctrl+C 停止服务")
	fmt.Println("========================================")
	fmt.Println()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "risk-decision-engine",
			"version": "0.2.0",
		})
	})

	// 决策接口
	r.POST("/api/v1/decision/execute", func(c *gin.Context) {
		var req struct {
			RequestID  string                 `json:"requestId"`
			BusinessID string                 `json:"businessId"`
			Data       map[string]interface{} `json:"data"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "1001",
				"message": "参数错误",
			})
			return
		}

		decisionID := "dec_" + uuid.NewString()[:8]
		fmt.Printf("[决策请求] requestId=%s, data=%+v\n", req.RequestID, req.Data)

		// 提取年龄
		age := extractAge(req.Data)
		fact := map[string]interface{}{
			"age": age,
		}

		// 执行规则
		decision := "APPROVE"
		reason := "通过"
		decisionCode := "APPROVE_001"

		if r, ok := ruleService.GetRule("R001"); ok {
			result, err := r.Execute(fact)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    "4001",
					"message": "规则执行失败",
				})
				return
			}

			if result.Action != nil {
				if result.Action.Result == "REJECT" {
					decision = "REJECT"
					decisionCode = "REJECT_R001"
					reason = result.Action.Reason
				} else {
					reason = "年龄符合要求"
				}
			}
		}

		fmt.Printf("[决策结果] decisionId=%s, decision=%s, reason=%s\n", decisionID, decision, reason)

		c.JSON(http.StatusOK, gin.H{
			"code":      "0000",
			"message":   "成功",
			"requestId": req.RequestID,
			"data": gin.H{
				"decisionId":     decisionID,
				"businessId":     req.BusinessID,
				"decision":       decision,
				"decisionCode":   decisionCode,
				"decisionReason": reason,
			},
		})
	})

	if err := r.Run("0.0.0.0:8080"); err != nil {
		fmt.Printf("启动失败: %v\n", err)
	}
}

func extractAge(data map[string]interface{}) int {
	if data == nil {
		return 0
	}

	// 尝试从 applicant.age 获取
	if applicant, ok := data["applicant"].(map[string]interface{}); ok {
		if age, ok := applicant["age"].(float64); ok {
			return int(age)
		}
		if age, ok := applicant["age"].(int); ok {
			return age
		}
	}

	// 尝试直接从 age 获取
	if age, ok := data["age"].(float64); ok {
		return int(age)
	}
	if age, ok := data["age"].(int); ok {
		return age
	}

	return 0
}
