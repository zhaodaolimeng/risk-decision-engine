package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"risk-decision-engine/internal/engine/rule"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	loadedRules  []*rule.SimpleRule
	rulesOnce    sync.Once
	ruleFilePath string
)

func initRules(configPath string) error {
	var initErr error
	rulesOnce.Do(func() {
		ruleFilePath = configPath
		fmt.Printf("加载规则配置文件: %s\n", configPath)

		rules, err := rule.LoadRulesFromFile(configPath)
		if err != nil {
			initErr = fmt.Errorf("load rules failed: %w", err)
			return
		}

		loadedRules = rules
		fmt.Printf("✓ 规则加载成功，共 %d 条规则\n", len(rules))
		for _, r := range rules {
			fmt.Printf("  - [%s] %s: %s\n", r.ID, r.Name, r.Expression)
		}
	})
	return initErr
}

func reloadRules(c *gin.Context) {
	if ruleFilePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "1001",
			"message": "规则文件未初始化",
		})
		return
	}

	// 重置并重新加载
	rulesOnce = sync.Once{}
	if err := initRules(ruleFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "4001",
			"message": "重载规则失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "0000",
		"message": "success",
		"data": gin.H{
			"ruleCount": len(loadedRules),
		},
	})
}

func executeConfigRuleDecision(c *gin.Context) {
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
	fmt.Printf("\n[配置规则决策请求] requestId=%s, decisionId=%s\n", req.RequestID, decisionID)

	// 构建事实数据
	fact := buildFactFromInput(req.Data)
	factJSON, _ := json.Marshal(fact)
	fmt.Printf("[事实数据] %s\n", string(factJSON))

	// 执行规则集
	ruleResults, err := rule.ExecuteRuleSet(loadedRules, fact)
	if err != nil {
		fmt.Printf("[决策错误] %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "4001",
			"message": "规则执行失败",
		})
		return
	}

	// 生成决策
	var decision, decisionCode, decisionReason string
	if ruleResults.AnyReject {
		decision = "REJECT"
		decisionCode = "REJECT_RULE"
		decisionReason = ruleResults.FirstRejectReason
	} else {
		decision = "APPROVE"
		decisionCode = "APPROVE_RULE"
		decisionReason = "所有规则通过"
	}

	fmt.Printf("[决策结果] decisionId=%s, decision=%s, code=%s, reason=%s\n",
		decisionID, decision, decisionCode, decisionReason)

	c.JSON(http.StatusOK, gin.H{
		"code":      "0000",
		"message":   "成功",
		"requestId": req.RequestID,
		"data": gin.H{
			"decisionId":     decisionID,
			"businessId":     req.BusinessID,
			"decision":       decision,
			"decisionCode":   decisionCode,
			"decisionReason": decisionReason,
			"ruleResults":    ruleResults,
		},
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

	// 处理嵌套字段（DS001.isBlacklist 风格）
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

func listRules(c *gin.Context) {
	type RuleInfo struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Expression string `json:"expression"`
	}

	var ruleList []RuleInfo
	for _, r := range loadedRules {
		ruleList = append(ruleList, RuleInfo{
			ID:         r.ID,
			Name:       r.Name,
			Expression: r.Expression,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "0000",
		"message": "success",
		"data": gin.H{
			"rules": ruleList,
		},
	})
}

func main() {
	fmt.Println("========================================")
	fmt.Println("   风险决策引擎 - 配置规则服务")
	fmt.Println("   从YAML文件动态加载规则")
	fmt.Println("========================================")
	fmt.Println()

	// 使用简单用例的规则配置
	configPath := "test/cases/simple/01-age-rule/config/rule.yaml"

	if err := initRules(configPath); err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("服务说明:")
	fmt.Println("  - 服务端口: 8080")
	fmt.Println()
	fmt.Println("API接口:")
	fmt.Println("  - 健康检查: GET /health")
	fmt.Println("  - 规则列表: GET /api/v1/rules")
	fmt.Println("  - 执行决策: POST /api/v1/decision/execute")
	fmt.Println("  - 重载规则: POST /api/v1/rules/reload")
	fmt.Println()
	fmt.Println("当前规则配置:")
	fmt.Println("  - 配置文件:", configPath)
	fmt.Println()
	fmt.Println("测试示例:")
	fmt.Println(`  年龄25岁(通过):
  {
    "requestId": "test001",
    "data": { "age": 25 }
  }`)
	fmt.Println(`  年龄18岁(拒绝):
  {
    "requestId": "test002",
    "data": { "age": 18 }
  }`)
	fmt.Println()
	fmt.Println("按 Ctrl+C 停止服务")
	fmt.Println("========================================")
	fmt.Println()

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "risk-decision-engine-config",
			"version": "0.6.0",
		})
	})

	router.GET("/api/v1/rules", listRules)
	router.POST("/api/v1/rules/reload", reloadRules)
	router.POST("/api/v1/decision/execute", executeConfigRuleDecision)

	fmt.Println("服务启动: 0.0.0.0:8080")
	if err := router.Run("0.0.0.0:8080"); err != nil {
		fmt.Printf("服务启动失败: %v\n", err)
	}
}
