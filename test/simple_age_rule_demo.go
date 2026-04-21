package main

import (
	"fmt"

	"risk-decision-engine/internal/engine/flow"
	"risk-decision-engine/internal/engine/rule"
	"risk-decision-engine/pkg/logger"
)

func main() {
	// 初始化日志
	logger.Init("info", "console", "")

	fmt.Println("=== 风险决策引擎 - 简单年龄规则测试 ===\n")

	// 1. 创建规则
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
		logger.Fatalf("Load rule failed: %v", err)
	}
	fmt.Println("✓ 规则加载成功:", r.Name)

	rules := map[string]*rule.Rule{
		"R001": r,
	}

	// 2. 创建决策流
	testFlow := &flow.Flow{
		FlowID:      "F001",
		Name:        "年龄准入决策流",
		StartNodeID: "START",
		Nodes: []flow.Node{
			{
				NodeID: "START",
				Type:   flow.NodeTypeStart,
			},
			{
				NodeID: "RULE_AGE",
				Type:   flow.NodeTypeRuleSet,
				Config: map[string]interface{}{
					"ruleIds": []interface{}{"R001"},
				},
			},
			{
				NodeID: "DECISION",
				Type:   flow.NodeTypeDecision,
				Config: map[string]interface{}{
					"decisionTable": map[string]interface{}{
						"rules": []interface{}{
							map[string]interface{}{
								"condition": "anyRuleReject == true",
								"result":    flow.DecisionReject,
								"reason":    "年龄不符合要求，需在21-60岁之间",
							},
							map[string]interface{}{
								"condition": "allRulesPass == true",
								"result":    flow.DecisionApprove,
								"reason":    "年龄符合要求",
							},
						},
					},
				},
			},
			{
				NodeID: "END",
				Type:   flow.NodeTypeEnd,
			},
		},
		Edges: []flow.Edge{
			{From: "START", To: "RULE_AGE"},
			{From: "RULE_AGE", To: "DECISION"},
			{From: "DECISION", To: "END"},
		},
	}
	testFlow.InitMaps()
	fmt.Println("✓ 决策流创建成功:", testFlow.Name)

	// 3. 测试用例
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
		fact := map[string]interface{}{
			"age": tc.age,
		}

		result, err := testFlow.Execute(fact, rules)
		if err != nil {
			fmt.Printf("  ✗ 执行失败: %v\n", err)
			continue
		}

		fmt.Printf("  ✓ 决策结果: %s\n", result.Decision)
		fmt.Printf("  ✓ 决策原因: %s\n", result.Reason)
		fmt.Printf("  ✓ 执行耗时: %v\n", result.ExecuteTime)
	}

	fmt.Println("\n=== 测试完成 ===")
}
