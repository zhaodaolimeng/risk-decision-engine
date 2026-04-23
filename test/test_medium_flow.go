package main

import (
	"fmt"

	"risk-decision-engine/internal/engine/flow"
	"risk-decision-engine/internal/engine/model"
	"risk-decision-engine/internal/engine/rule"
)

// MockModelClient Mock模型客户端
type MockModelClient struct {
	mockRiskScore int
}

func NewMockModelClient(riskScore int) *MockModelClient {
	return &MockModelClient{mockRiskScore: riskScore}
}

func (m *MockModelClient) Predict(modelID string, req *model.ModelRequest) (*model.ModelResult, error) {
	fmt.Printf("[Mock模型] modelId=%s, contractId=%s, riskScore=%d\n", modelID, req.ContractID, m.mockRiskScore)

	var riskLevel string
	switch {
	case m.mockRiskScore >= 700:
		riskLevel = "LOW"
	case m.mockRiskScore >= 600:
		riskLevel = "MEDIUM"
	default:
		riskLevel = "HIGH"
	}

	return &model.ModelResult{
		ModelID:            modelID,
		DefaultProbability: 0.1,
		RiskScore:          m.mockRiskScore,
		RiskLevel:          riskLevel,
	}, nil
}

func main() {
	fmt.Println("========================================")
	fmt.Println("   中等用例 - 规则+模型联合决策 - 自测")
	fmt.Println("========================================")
	fmt.Println()

	// 创建规则
	ageRule := rule.NewAgeRule()
	incomeRule := rule.NewIncomeRule()
	rules := []*rule.SimpleRule{ageRule, incomeRule}
	fmt.Printf("✓ 规则加载成功: %s, %s\n", ageRule.Name, incomeRule.Name)
	fmt.Printf("  条件: %s; %s\n", ageRule.Expression, incomeRule.Expression)
	fmt.Println()

	// 测试用例
	testCases := []struct {
		name           string
		age            int
		monthlyIncome  int
		mockRiskScore  int
		expectedDecision string
	}{
		{"规则通过 + 模型低分(750) -> 通过", 34, 30000, 750, "APPROVE"},
		{"规则通过 + 模型中分(650) -> 人工复核", 34, 30000, 650, "MANUAL"},
		{"规则通过 + 模型高分(550) -> 拒绝", 34, 30000, 550, "REJECT"},
		{"年龄规则拒绝(18岁)", 18, 30000, 750, "REJECT"},
		{"收入规则拒绝(4000元)", 30, 4000, 750, "REJECT"},
	}

	fmt.Println("开始测试:")
	fmt.Println("----------------")

	allPassed := true
	for _, tc := range testCases {
		fmt.Printf("\n测试: %s\n", tc.name)

		// 创建mock模型客户端
		mockClient := NewMockModelClient(tc.mockRiskScore)

		// 创建决策流引擎
		engine := &flow.SimpleFlowEngine{
			rules:       rules,
			modelClient: mockClient,
		}

		// 准备输入数据
		input := map[string]interface{}{
			"contractId": "CTR_TEST_001",
			"applicant": map[string]interface{}{
				"age":           tc.age,
				"monthlyIncome": tc.monthlyIncome,
			},
		}

		// 执行决策流
		ctx := &flow.DecisionContext{
			Input: input,
		}

		err := engine.Execute(ctx)
		if err != nil {
			fmt.Printf("   ✗ 执行失败: %v\n", err)
			allPassed = false
			continue
		}

		fmt.Printf("   原因: %s\n", ctx.DecisionReason)
		fmt.Printf("   决策码: %s\n", ctx.DecisionCode)

		if ctx.Decision == tc.expectedDecision {
			fmt.Printf("   ✓ 决策: %s (期望: %s)\n", ctx.Decision, tc.expectedDecision)
		} else {
			fmt.Printf("   ✗ 决策: %s (期望: %s)\n", ctx.Decision, tc.expectedDecision)
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
