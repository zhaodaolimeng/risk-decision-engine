package main

import (
	"fmt"

	"risk-decision-engine/internal/engine/rule"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("   规则配置文件加载 - 自测")
	fmt.Println("========================================")
	fmt.Println()

	// 测试1: 简单用例配置
	fmt.Println("测试1: 简单用例规则配置")
	fmt.Println("--------------------------------")
	testSimpleConfig()
	fmt.Println()

	// 测试2: 中等用例配置
	fmt.Println("测试2: 中等用例规则配置")
	fmt.Println("--------------------------------")
	testMediumConfig()
	fmt.Println()

	// 测试3: 复杂用例配置
	fmt.Println("测试3: 复杂用例规则配置")
	fmt.Println("--------------------------------")
	testComplexConfig()
	fmt.Println()

	fmt.Println("================================")
	fmt.Println("✓ 所有测试完成!")
}

func testSimpleConfig() {
	configPath := "test/cases/simple/01-age-rule/config/rule.yaml"

	// 加载配置
	config, err := rule.LoadRuleConfigFromFile(configPath)
	if err != nil {
		fmt.Printf("✗ 加载配置失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 配置加载成功，规则数: %d\n", len(config.Rules))

	// 转换为规则
	rules, err := rule.LoadRulesFromConfig(config)
	if err != nil {
		fmt.Printf("✗ 转换规则失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 规则转换成功，加载规则数: %d\n", len(rules))

	// 打印规则
	for _, r := range rules {
		fmt.Printf("  - [%s] %s\n", r.ID, r.Name)
		fmt.Printf("    表达式: %s\n", r.Expression)
	}

	// 测试规则执行
	testRuleExecution(rules, map[string]interface{}{"age": 25}, "PASS")
	testRuleExecution(rules, map[string]interface{}{"age": 18}, "REJECT")
	testRuleExecution(rules, map[string]interface{}{"age": 65}, "REJECT")
}

func testMediumConfig() {
	configPath := "test/cases/medium/01-rule-and-model/config/rule.yaml"

	// 加载配置
	config, err := rule.LoadRuleConfigFromFile(configPath)
	if err != nil {
		fmt.Printf("✗ 加载配置失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 配置加载成功，规则数: %d\n", len(config.Rules))

	// 转换为规则
	rules, err := rule.LoadRulesFromConfig(config)
	if err != nil {
		fmt.Printf("✗ 转换规则失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 规则转换成功，加载规则数: %d\n", len(rules))

	// 打印规则
	for _, r := range rules {
		fmt.Printf("  - [%s] %s\n", r.ID, r.Name)
		fmt.Printf("    表达式: %s\n", r.Expression)
	}

	// 测试规则执行
	testRuleExecution(rules, map[string]interface{}{"age": 30, "monthlyIncome": 10000}, "PASS")
	testRuleExecution(rules, map[string]interface{}{"age": 30, "monthlyIncome": 4000}, "REJECT")
	testRuleExecution(rules, map[string]interface{}{"age": 18, "monthlyIncome": 10000}, "REJECT")
}

func testComplexConfig() {
	configPath := "test/cases/complex/01-datasource-rule-model/config/rule.yaml"

	// 加载配置
	config, err := rule.LoadRuleConfigFromFile(configPath)
	if err != nil {
		fmt.Printf("✗ 加载配置失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 配置加载成功，规则数: %d\n", len(config.Rules))

	// 转换为规则
	rules, err := rule.LoadRulesFromConfig(config)
	if err != nil {
		fmt.Printf("✗ 转换规则失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 规则转换成功，加载规则数: %d\n", len(rules))

	// 打印规则
	for _, r := range rules {
		fmt.Printf("  - [%s] %s\n", r.ID, r.Name)
		fmt.Printf("    表达式: %s\n", r.Expression)
	}

	// 测试规则执行（需要嵌套数据结构）
	fact1 := map[string]interface{}{
		"DS001.isBlacklist": false,
		"input.applicant.age": 30,
		"DS003.rejectCount7D": 1,
		"DS003.multiScore": 80,
	}
	testRuleExecution(rules, fact1, "PASS")

	fact2 := map[string]interface{}{
		"DS001.isBlacklist": true,
		"input.applicant.age": 30,
		"DS003.rejectCount7D": 1,
		"DS003.multiScore": 80,
	}
	testRuleExecution(rules, fact2, "REJECT")
}

func testRuleExecution(rules []*rule.SimpleRule, fact map[string]interface{}, expected string) {
	result, err := rule.ExecuteRuleSet(rules, fact)
	if err != nil {
		fmt.Printf("  ✗ 执行失败 (fact=%v): %v\n", fact, err)
		return
	}

	var actual string
	if result.AnyReject {
		actual = "REJECT"
	} else {
		actual = "PASS"
	}

	if actual == expected {
		fmt.Printf("  ✓ 测试通过 (fact=%v) → %s\n", fact, actual)
	} else {
		fmt.Printf("  ✗ 测试失败 (fact=%v) → %s (期望: %s)\n", fact, actual, expected)
	}
}
