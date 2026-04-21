package test

import (
	"os"
	"path/filepath"
	"testing"

	"risk-decision-engine/internal/engine/flow"
	"risk-decision-engine/internal/engine/rule"
	"risk-decision-engine/pkg/logger"

	"github.com/stretchr/testify/assert"
)

func init() {
	// 初始化日志
	logger.Init("debug", "console", "")
}

func TestSimpleAgeRule(t *testing.T) {
	// 加载规则
	rulePath := filepath.Join("cases", "simple", "01-age-rule", "config", "rule.yaml")
	ruleData, err := os.ReadFile(rulePath)
	assert.NoError(t, err)

	// 简单的YAML解析为JSON（这里简化处理，直接写硬编码规则用于测试）
	rules := loadTestRules(t)

	// 加载决策流（简化版，硬编码用于测试）
	testFlow := createTestFlow()

	// 测试场景1: 年龄20岁 - 拒绝
	t.Run("Age20_Reject", func(t *testing.T) {
		fact := map[string]interface{}{
			"age": 20,
		}

		result, err := testFlow.Execute(fact, rules)
		assert.NoError(t, err)
		assert.Equal(t, flow.DecisionReject, result.Decision)
		assert.Contains(t, result.Reason, "年龄不符合")
		t.Logf("Result: %+v", result)
	})

	// 测试场景2: 年龄25岁 - 通过
	t.Run("Age25_Approve", func(t *testing.T) {
		fact := map[string]interface{}{
			"age": 25,
		}

		result, err := testFlow.Execute(fact, rules)
		assert.NoError(t, err)
		assert.Equal(t, flow.DecisionApprove, result.Decision)
		t.Logf("Result: %+v", result)
	})

	// 测试场景3: 年龄60岁 - 通过
	t.Run("Age60_Approve", func(t *testing.T) {
		fact := map[string]interface{}{
			"age": 60,
		}

		result, err := testFlow.Execute(fact, rules)
		assert.NoError(t, err)
		assert.Equal(t, flow.DecisionApprove, result.Decision)
		t.Logf("Result: %+v", result)
	})

	// 测试场景4: 年龄61岁 - 拒绝
	t.Run("Age61_Reject", func(t *testing.T) {
		fact := map[string]interface{}{
			"age": 61,
		}

		result, err := testFlow.Execute(fact, rules)
		assert.NoError(t, err)
		assert.Equal(t, flow.DecisionReject, result.Decision)
		t.Logf("Result: %+v", result)
	})
}

func loadTestRules(t *testing.T) map[string]*rule.Rule {
	// 创建测试规则
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
	assert.NoError(t, err)

	return map[string]*rule.Rule{
		"R001": r,
	}
}

func createTestFlow() *flow.Flow {
	return &flow.Flow{
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
}
