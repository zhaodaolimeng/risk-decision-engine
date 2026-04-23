package sandbox

import (
	"fmt"
	"sync"
	"time"

	"risk-decision-engine/internal/engine/rule"
)

// ConfigurableExecutor 可配置的决策执行器
type ConfigurableExecutor struct {
	mu       sync.RWMutex
	rules    []*rule.SimpleRule
	configPath string
}

// NewConfigurableExecutor 创建可配置执行器
func NewConfigurableExecutor(defaultConfigPath string) (*ConfigurableExecutor, error) {
	e := &ConfigurableExecutor{
		configPath: defaultConfigPath,
	}
	if defaultConfigPath != "" {
		if err := e.LoadConfig(defaultConfigPath); err != nil {
			return nil, err
		}
	}
	return e, nil
}

// LoadConfig 加载规则配置
func (e *ConfigurableExecutor) LoadConfig(configPath string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	rules, err := rule.LoadRulesFromFile(configPath)
	if err != nil {
		return fmt.Errorf("load rules: %w", err)
	}

	e.rules = rules
	e.configPath = configPath
	return nil
}

// GetRules 获取当前规则
func (e *ConfigurableExecutor) GetRules() []*rule.SimpleRule {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.rules
}

// Execute 执行决策（实现DecisionExecutor接口）
func (e *ConfigurableExecutor) Execute(
	requestID, businessID string,
	data map[string]interface{},
) (
	decision, decisionCode, decisionReason string,
	ruleResults, modelResult interface{},
	duration time.Duration,
	err error,
) {
	start := time.Now()

	e.mu.RLock()
	rules := e.rules
	e.mu.RUnlock()

	if len(rules) == 0 {
		duration = time.Since(start)
		return "", "", "", nil, nil, duration, fmt.Errorf("no rules loaded")
	}

	// 构建事实数据
	fact := buildFactFromInput(data)

	// 执行规则集
	results, err := rule.ExecuteRuleSet(rules, fact)
	if err != nil {
		duration = time.Since(start)
		return "", "", "", nil, nil, duration, err
	}

	// 生成决策
	if results.AnyReject {
		decision = "REJECT"
		decisionCode = "REJECT_RULE"
		decisionReason = results.FirstRejectReason
	} else {
		decision = "APPROVE"
		decisionCode = "APPROVE_RULE"
		decisionReason = "所有规则通过"
	}

	duration = time.Since(start)
	return decision, decisionCode, decisionReason, results, nil, duration, nil
}

// buildFactFromInput 从输入构建事实数据
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
