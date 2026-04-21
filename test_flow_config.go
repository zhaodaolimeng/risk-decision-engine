package main

import (
	"fmt"

	"risk-decision-engine/internal/engine/flow"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("   决策流配置文件加载 - 自测")
	fmt.Println("========================================")
	fmt.Println()

	// 测试1: 简单用例配置
	fmt.Println("测试1: 简单用例决策流配置")
	fmt.Println("--------------------------------")
	testSimpleFlowConfig()
	fmt.Println()

	// 测试2: 中等用例配置
	fmt.Println("测试2: 中等用例决策流配置")
	fmt.Println("--------------------------------")
	testMediumFlowConfig()
	fmt.Println()

	// 测试3: 复杂用例配置
	fmt.Println("测试3: 复杂用例决策流配置")
	fmt.Println("--------------------------------")
	testComplexFlowConfig()
	fmt.Println()

	fmt.Println("================================")
	fmt.Println("✓ 所有测试完成!")
}

func testSimpleFlowConfig() {
	configPath := "test/cases/simple/01-age-rule/config/flow.yaml"

	// 加载配置
	config, err := flow.LoadFlowConfigFromFile(configPath)
	if err != nil {
		fmt.Printf("✗ 加载配置失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 配置加载成功\n")

	// 验证配置
	if err := config.Validate(); err != nil {
		fmt.Printf("✗ 配置验证失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 配置验证通过\n")

	// 打印配置信息
	fmt.Printf("  Flow ID: %s\n", config.FlowID)
	fmt.Printf("  Name: %s\n", config.Name)
	fmt.Printf("  Status: %s\n", config.Status)
	fmt.Printf("  Nodes: %d\n", len(config.Nodes))
	fmt.Printf("  Edges: %d\n", len(config.Edges))

	// 打印节点
	fmt.Println("  Nodes:")
	for _, node := range config.Nodes {
		fmt.Printf("    - [%s] %s", node.NodeID, node.Type)
		if len(node.RuleIds) > 0 {
			fmt.Printf(" (rules: %v)", node.RuleIds)
		}
		if node.ModelID != "" {
			fmt.Printf(" (model: %s)", node.ModelID)
		}
		fmt.Println()
	}

	if config.IsActive() {
		fmt.Println("✓ 配置状态: ACTIVE")
	}
}

func testMediumFlowConfig() {
	configPath := "test/cases/medium/01-rule-and-model/config/flow.yaml"

	// 加载配置
	config, err := flow.LoadFlowConfigFromFile(configPath)
	if err != nil {
		fmt.Printf("✗ 加载配置失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 配置加载成功\n")

	// 验证配置
	if err := config.Validate(); err != nil {
		fmt.Printf("✗ 配置验证失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 配置验证通过\n")

	// 打印配置信息
	fmt.Printf("  Flow ID: %s\n", config.FlowID)
	fmt.Printf("  Name: %s\n", config.Name)
	fmt.Printf("  Status: %s\n", config.Status)
	fmt.Printf("  Nodes: %d\n", len(config.Nodes))
	fmt.Printf("  Edges: %d\n", len(config.Edges))

	// 检查决策表
	for _, node := range config.Nodes {
		if node.Type == "DECISION" && node.DecisionTable != nil {
			fmt.Printf("  决策表规则数: %d\n", len(node.DecisionTable.Rules))
		}
	}

	if config.IsActive() {
		fmt.Println("✓ 配置状态: ACTIVE")
	}
}

func testComplexFlowConfig() {
	configPath := "test/cases/complex/01-datasource-rule-model/config/flow.yaml"

	// 加载配置
	config, err := flow.LoadFlowConfigFromFile(configPath)
	if err != nil {
		fmt.Printf("✗ 加载配置失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 配置加载成功\n")

	// 验证配置
	if err := config.Validate(); err != nil {
		fmt.Printf("✗ 配置验证失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 配置验证通过\n")

	// 打印配置信息
	fmt.Printf("  Flow ID: %s\n", config.FlowID)
	fmt.Printf("  Name: %s\n", config.Name)
	fmt.Printf("  Status: %s\n", config.Status)
	fmt.Printf("  Nodes: %d\n", len(config.Nodes))
	fmt.Printf("  Edges: %d\n", len(config.Edges))

	// 检查预加载配置
	if config.Preload != nil {
		fmt.Printf("  预加载数据源: %v\n", config.Preload.DatasourceIds)
		fmt.Printf("  并行加载: %t\n", config.Preload.Parallel)
		fmt.Printf("  超时: %s\n", config.Preload.Timeout)
	}

	if config.IsActive() {
		fmt.Println("✓ 配置状态: ACTIVE")
	}
}
