package main

import (
	"fmt"

	"risk-decision-engine/internal/engine/datasource"
	"risk-decision-engine/internal/engine/flow"
	"risk-decision-engine/internal/engine/model"
	"risk-decision-engine/internal/engine/rule"
)

// MockDataSourceClient Mock数据源客户端
type MockDataSourceClient struct {
	data map[string]interface{}
}

func NewMockDataSourceClient(data map[string]interface{}) *MockDataSourceClient {
	return &MockDataSourceClient{data: data}
}

func (m *MockDataSourceClient) Fetch(dsID string, req *datasource.DataSourceRequest) (*datasource.DataSourceResult, error) {
	fmt.Printf("[Mock数据源] dsId=%s, userId=%s\n", dsID, req.UserID)
	return &datasource.DataSourceResult{
		DataSourceID: dsID,
		Data:         m.data,
	}, nil
}

// MockComplexModelClient Mock模型客户端
type MockComplexModelClient struct {
	mockRiskScore int
}

func NewMockComplexModelClient(riskScore int) *MockComplexModelClient {
	return &MockComplexModelClient{mockRiskScore: riskScore}
}

func (m *MockComplexModelClient) Predict(modelID string, req *model.ModelRequest) (*model.ModelResult, error) {
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
	fmt.Println("   复杂用例 - 数据源+规则+模型 - 自测")
	fmt.Println("========================================")
	fmt.Println()

	// 测试用例
	testCases := []struct {
		name             string
		userData         map[string]interface{}
		creditData       map[string]interface{}
		multiData        map[string]interface{}
		mockRiskScore    int
		expectedDecision string
	}{
		{
			name:             "黑名单用户 -> 直接拒绝",
			userData:         map[string]interface{}{"isBlacklist": true},
			creditData:       map[string]interface{}{"creditScore": 750},
			multiData:        map[string]interface{}{"multiQueryCount7d": 1},
			mockRiskScore:    750,
			expectedDecision: "REJECT",
		},
		{
			name:             "正常用户+低多头+低风险 -> 通过",
			userData:         map[string]interface{}{"isBlacklist": false, "age": 35},
			creditData:       map[string]interface{}{"creditScore": 750},
			multiData:        map[string]interface{}{"multiQueryCount7d": 1},
			mockRiskScore:    750,
			expectedDecision: "APPROVE",
		},
		{
			name:             "高多头+中风险 -> 人工复核",
			userData:         map[string]interface{}{"isBlacklist": false, "age": 35},
			creditData:       map[string]interface{}{"creditScore": 700},
			multiData:        map[string]interface{}{"multiQueryCount7d": 8},
			mockRiskScore:    650,
			expectedDecision: "MANUAL",
		},
		{
			name:             "征信不良+高风险 -> 拒绝",
			userData:         map[string]interface{}{"isBlacklist": false, "age": 35},
			creditData:       map[string]interface{}{"creditScore": 500, "hasOverdue": true},
			multiData:        map[string]interface{}{"multiQueryCount7d": 4},
			mockRiskScore:    550,
			expectedDecision: "REJECT",
		},
		{
			name:             "年龄不足 -> 规则拒绝",
			userData:         map[string]interface{}{"isBlacklist": false, "age": 18},
			creditData:       map[string]interface{}{"creditScore": 750},
			multiData:        map[string]interface{}{"multiQueryCount7d": 1},
			mockRiskScore:    750,
			expectedDecision: "REJECT",
		},
	}

	fmt.Println("开始测试:")
	fmt.Println("----------------")

	allPassed := true
	for _, tc := range testCases {
		fmt.Printf("\n测试: %s\n", tc.name)

		// 创建Mock数据源客户端
		dsClient1 := NewMockDataSourceClient(tc.userData)
		dsClient2 := NewMockDataSourceClient(tc.creditData)
		dsClient3 := NewMockDataSourceClient(tc.multiData)

		// 配置数据源
		dataSources := []*flow.DataSourceConfig{
			{ID: "DS001", Client: dsClient1, Required: true},
			{ID: "DS002", Client: dsClient2, Required: true},
			{ID: "DS003", Client: dsClient3, Required: true},
		}

		// 创建规则
		blacklistRule := rule.NewBlacklistRule()
		ageRule := rule.NewAgeRule()
		multiRule := rule.NewMultiQueryRule()
		rules := []*rule.SimpleRule{blacklistRule, ageRule, multiRule}

		// 创建Mock模型客户端
		mockModelClient := NewMockComplexModelClient(tc.mockRiskScore)

		// 创建复杂决策流引擎
		engine := flow.NewComplexFlowEngine(dataSources, rules, mockModelClient)

		// 准备输入数据
		input := map[string]interface{}{
			"contractId": "CTR_TEST_001",
			"userId":     "USER_TEST_001",
			"applicant": map[string]interface{}{
				"age": tc.userData["age"],
			},
		}

		// 执行决策流
		ctx := &flow.ComplexDecisionContext{
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
