package main

import (
	"fmt"

	"risk-decision-engine/internal/engine/rule"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("   规则引擎 - 自测")
	fmt.Println("========================================")
	fmt.Println()

	// 创建规则
	ageRule := rule.NewAgeRule()
	fmt.Printf("✓ 规则加载成功: %s\n", ageRule.Name)
	fmt.Printf("  条件: %s\n", ageRule.Expression)
	fmt.Println()

	// 测试用例
	testCases := []struct {
		name     string
		age      int
		expected string
	}{
		{"年龄 20 岁", 20, "REJECT"},
		{"年龄 25 岁", 25, "APPROVE"},
		{"年龄 60 岁", 60, "APPROVE"},
		{"年龄 61 岁", 61, "REJECT"},
	}

	fmt.Println("开始测试:")
	fmt.Println("----------------")

	allPassed := true
	for _, tc := range testCases {
		fmt.Printf("\n测试: %s\n", tc.name)

		fact := map[string]interface{}{
			"age": tc.age,
		}

		result, err := ageRule.Execute(fact)
		if err != nil {
			fmt.Printf("   ✗ 执行失败: %v\n", err)
			allPassed = false
			continue
		}

		var decision string
		if result.Result == "REJECT" {
			decision = "REJECT"
		} else {
			decision = "APPROVE"
		}

		if result.Reason != "" {
			fmt.Printf("   原因: %s\n", result.Reason)
		}

		if decision == tc.expected {
			fmt.Printf("   ✓ 决策: %s (期望: %s)\n", decision, tc.expected)
		} else {
			fmt.Printf("   ✗ 决策: %s (期望: %s)\n", decision, tc.expected)
			allPassed = false
		}
	}

	fmt.Println("\n----------------")
	if allPassed {
		fmt.Println("✓ 所有测试通过!")
	} else {
		fmt.Println("✗ 部分测试失败")
	}
	fmt.Println("================================")
}
