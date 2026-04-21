package rule

import (
	"fmt"
	"time"

	"github.com/expr-lang/expr"
)

// SimpleRule 简单规则
type SimpleRule struct {
	ID         string
	Name       string
	Expression string
	PassAction *Action
	FailAction *Action
}

// Action 动作
type Action struct {
	Result string
	Reason string
}

// SimpleRuleResult 规则执行结果
type SimpleRuleResult struct {
	RuleID string
	Pass   bool
	Result string
	Reason string
}

// RuleSetResult 规则集执行结果
type RuleSetResult struct {
	AllPass        bool
	AnyReject      bool
	Results        []*SimpleRuleResult
	FirstRejectReason string
}

// NewAgeRule 创建年龄规则
func NewAgeRule() *SimpleRule {
	return &SimpleRule{
		ID:         "R001",
		Name:       "年龄准入规则",
		Expression: "age >= 21 && age <= 60",
		PassAction: &Action{
			Result: "PASS",
		},
		FailAction: &Action{
			Result: "REJECT",
			Reason: "年龄不符合要求，需在21-60岁之间",
		},
	}
}

// NewIncomeRule 创建收入规则
func NewIncomeRule() *SimpleRule {
	return &SimpleRule{
		ID:         "R002",
		Name:       "收入准入规则",
		Expression: "monthlyIncome >= 5000",
		PassAction: &Action{
			Result: "PASS",
		},
		FailAction: &Action{
			Result: "REJECT",
			Reason: "月收入不足，需达到5000元",
		},
	}
}

// NewBlacklistRule 创建黑名单规则
func NewBlacklistRule() *SimpleRule {
	return &SimpleRule{
		ID:         "R003",
		Name:       "黑名单规则",
		Expression: "isBlacklist != true",
		PassAction: &Action{
			Result: "PASS",
		},
		FailAction: &Action{
			Result: "REJECT",
			Reason: "用户在黑名单中",
		},
	}
}

// NewMultiQueryRule 创建多头查询规则
func NewMultiQueryRule() *SimpleRule {
	return &SimpleRule{
		ID:         "R004",
		Name:       "多头查询规则",
		Expression: "multiQueryCount7d == nil || multiQueryCount7d <= 10",
		PassAction: &Action{
			Result: "PASS",
		},
		FailAction: &Action{
			Result: "REJECT",
			Reason: "近7天多头查询次数过多",
		},
	}
}

// ExecuteRuleSet 执行规则集
func ExecuteRuleSet(rules []*SimpleRule, fact map[string]interface{}) (*RuleSetResult, error) {
	result := &RuleSetResult{
		AllPass:   true,
		AnyReject: false,
		Results:   make([]*SimpleRuleResult, 0, len(rules)),
	}

	for _, rule := range rules {
		ruleResult, err := rule.Execute(fact)
		if err != nil {
			return nil, err
		}

		result.Results = append(result.Results, ruleResult)

		if !ruleResult.Pass {
			result.AllPass = false
			result.AnyReject = true
			if result.FirstRejectReason == "" {
				result.FirstRejectReason = ruleResult.Reason
			}
		}
	}

	return result, nil
}

// Execute 执行规则
func (r *SimpleRule) Execute(fact map[string]interface{}) (*SimpleRuleResult, error) {
	start := time.Now()
	defer func() {
		fmt.Printf("[规则执行] %s 耗时: %v\n", r.Name, time.Since(start))
	}()

	output, err := expr.Eval(r.Expression, fact)
	if err != nil {
		return nil, fmt.Errorf("eval expression: %w", err)
	}

	pass, ok := output.(bool)
	if !ok {
		return nil, fmt.Errorf("expression result not boolean: %v", output)
	}

	result := &SimpleRuleResult{
		RuleID: r.ID,
		Pass:   pass,
	}

	if pass {
		if r.PassAction != nil {
			result.Result = r.PassAction.Result
		}
	} else {
		if r.FailAction != nil {
			result.Result = r.FailAction.Result
			result.Reason = r.FailAction.Reason
		}
	}

	fmt.Printf("[规则执行] %s, age=%v, pass=%v, result=%s\n",
		r.Name, fact["age"], pass, result.Result)

	return result, nil
}
