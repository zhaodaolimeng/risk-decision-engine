package rule

import (
	"encoding/json"
	"fmt"
	"time"

	"risk-decision-engine/internal/engine/expression"
	"risk-decision-engine/pkg/logger"
)

// RuleType 规则类型
type RuleType string

const (
	RuleTypeBoolean RuleType = "BOOLEAN"
	RuleTypeNumeric RuleType = "NUMERIC"
)

// RuleStatus 规则状态
type RuleStatus string

const (
	RuleStatusDraft    RuleStatus = "DRAFT"
	RuleStatusActive   RuleStatus = "ACTIVE"
	RuleStatusDisabled RuleStatus = "DISABLED"
)

// Rule 规则定义
type Rule struct {
	RuleID      string                 `json:"ruleId"`
	Version     string                 `json:"version"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        RuleType               `json:"type"`
	Condition   map[string]interface{} `json:"condition"`
	Actions     map[string]interface{} `json:"actions"`
	Priority    int                    `json:"priority"`
	Status      RuleStatus             `json:"status"`

	// 编译后的表达式
	compiledCondition string
}

// ActionResult 动作结果
type ActionResult struct {
	Result string                 `json:"result"`
	Reason string                 `json:"reason,omitempty"`
	Output map[string]interface{} `json:"output,omitempty"`
}

// RuleResult 规则执行结果
type RuleResult struct {
	RuleID      string        `json:"ruleId"`
	RuleName    string        `json:"ruleName"`
	Pass        bool          `json:"pass"`
	Hit         bool          `json:"hit"`
	Action      *ActionResult `json:"action,omitempty"`
	ExecuteTime time.Duration `json:"executeTime,omitempty"`
}

// LoadRuleFromJSON 从JSON加载规则
func LoadRuleFromJSON(data []byte) (*Rule, error) {
	var rule Rule
	if err := json.Unmarshal(data, &rule); err != nil {
		return nil, fmt.Errorf("unmarshal rule: %w", err)
	}

	if err := rule.Compile(); err != nil {
		return nil, fmt.Errorf("compile rule: %w", err)
	}

	return &rule, nil
}

// Compile 编译规则条件
func (r *Rule) Compile() error {
	exprStr, err := buildExprFromCondition(r.Condition)
	if err != nil {
		return err
	}
	r.compiledCondition = exprStr
	logger.Debugf("Rule %s compiled: %s", r.RuleID, exprStr)
	return nil
}

// Execute 执行规则
func (r *Rule) Execute(fact map[string]interface{}) (*RuleResult, error) {
	start := time.Now()
	result := &RuleResult{
		RuleID:   r.RuleID,
		RuleName: r.Name,
	}

	if r.Status != RuleStatusActive {
		result.Hit = false
		result.Pass = false
		result.ExecuteTime = time.Since(start)
		return result, nil
	}

	// 执行条件
	pass, err := expression.EvaluateBoolean(r.compiledCondition, fact)
	if err != nil {
		return nil, fmt.Errorf("evaluate condition: %w", err)
	}

	result.Hit = true
	result.Pass = pass

	// 执行动作
	action, err := r.executeAction(pass)
	if err != nil {
		return nil, fmt.Errorf("execute action: %w", err)
	}
	result.Action = action

	result.ExecuteTime = time.Since(start)
	return result, nil
}

func (r *Rule) executeAction(pass bool) (*ActionResult, error) {
	key := "true"
	if !pass {
		key = "false"
	}

	actionData, ok := r.Actions[key]
	if !ok {
		return &ActionResult{
			Result: "PASS",
		}, nil
	}

	actionMap, ok := actionData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid action format")
	}

	result := &ActionResult{}
	if res, ok := actionMap["result"]; ok {
		result.Result = fmt.Sprintf("%v", res)
	}
	if reason, ok := actionMap["reason"]; ok {
		result.Reason = fmt.Sprintf("%v", reason)
	}
	if output, ok := actionMap["output"]; ok {
		if outputMap, ok := output.(map[string]interface{}); ok {
			result.Output = outputMap
		}
	}

	return result, nil
}

func buildExprFromCondition(condition map[string]interface{}) (string, error) {
	if condition == nil {
		return "true", nil
	}

	// 检查是否是复合条件
	if op, ok := condition["operator"]; ok {
		operator := fmt.Sprintf("%v", op)
		exprs, ok := condition["expressions"].([]interface{})
		if !ok {
			return "", fmt.Errorf("invalid compound condition format")
		}

		var subExprs []string
		for _, e := range exprs {
			exprMap, ok := e.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("invalid expression format")
			}
			subExpr, err := buildExprFromCondition(exprMap)
			if err != nil {
				return "", err
			}
			subExprs = append(subExprs, "("+subExpr+")")
		}

		if operator == "AND" {
			return joinExprs(subExprs, "&&"), nil
		} else if operator == "OR" {
			return joinExprs(subExprs, "||"), nil
		}
	}

	// 简单条件
	field, _ := condition["field"].(string)
	operator, _ := condition["operator"].(string)
	value := condition["value"]

	if field == "" || operator == "" {
		return "", fmt.Errorf("invalid simple condition format")
	}

	return expression.BuildConditionExpr(field, operator, value), nil
}

func joinExprs(exprs []string, op string) string {
	if len(exprs) == 0 {
		return "true"
	}
	result := exprs[0]
	for i := 1; i < len(exprs); i++ {
		result += " " + op + " " + exprs[i]
	}
	return result
}
