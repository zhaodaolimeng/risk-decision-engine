package main

import (
	"fmt"

	"github.com/expr-lang/expr"
)

type Rule struct {
	ID          string
	Name        string
	Condition   string
	PassResult  string
	FailResult  string
	FailReason  string
}

type RuleResult struct {
	RuleID string
	Pass   bool
	Result string
	Reason string
}

func (r *Rule) Execute(env map[string]interface{}) (*RuleResult, error) {
	output, err := expr.Eval(r.Condition, env)
	if err != nil {
		return nil, err
	}

	pass, ok := output.(bool)
	if !ok {
		return nil, fmt.Errorf("condition result not boolean")
	}

	result := &RuleResult{
		RuleID: r.ID,
		Pass:   pass,
	}

	if pass {
		result.Result = r.PassResult
	} else {
		result.Result = r.FailResult
		result.Reason = r.FailReason
	}

	return result, nil
}

func main() {
	fmt.Println("=== 风险决策引擎 - 简单规则测试 ===\n")

	// 创建规则
	ageRule := &Rule{
		ID:         "R001",
		Name:       "年龄准入规则",
		Condition:  "age >= 21 && age <= 60",
		PassResult: "PASS",
		FailResult: "REJECT",
		FailReason: "年龄不符合要求，需在21-60岁之间",
	}

	fmt.Println("✓ 规则加载成功:", ageRule.Name)
	fmt.Println("  条件:", ageRule.Condition)

	// 测试用例
	testCases := []struct {
		name string
		age  int
	}{
		{"年龄 20 岁", 20},
		{"年龄 25 岁", 25},
		{"年龄 60 岁", 60},
		{"年龄 61 岁", 61},
	}

	fmt.Println("\n=== 开始测试 ===")
	for _, tc := range testCases {
		fmt.Printf("\n测试: %s\n", tc.name)
		env := map[string]interface{}{
			"age": tc.age,
		}

		result, err := ageRule.Execute(env)
		if err != nil {
			fmt.Printf("  ✗ 执行失败: %v\n", err)
			continue
		}

		fmt.Printf("  ✓ 规则ID: %s\n", result.RuleID)
		fmt.Printf("  ✓ 是否通过: %v\n", result.Pass)
		fmt.Printf("  ✓ 结果: %s\n", result.Result)
		if result.Reason != "" {
			fmt.Printf("  ✓ 原因: %s\n", result.Reason)
		}
	}

	fmt.Println("\n=== 测试完成 ===")
}
