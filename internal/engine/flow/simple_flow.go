package flow

import (
	"fmt"

	"risk-decision-engine/internal/engine/model"
	"risk-decision-engine/internal/engine/rule"
)

// DecisionContext 决策上下文
type DecisionContext struct {
	Input          map[string]interface{}
	RuleResults    *rule.RuleSetResult
	ModelResult    *model.ModelResult
	Decision       string
	DecisionReason string
	DecisionCode   string
}

// SimpleFlowEngine 简化版决策流引擎
type SimpleFlowEngine struct {
	rules       []*rule.SimpleRule
	modelClient model.ModelService
}

// NewSimpleFlowEngine 创建简化版决策流引擎
func NewSimpleFlowEngine(rules []*rule.SimpleRule, modelClient model.ModelService) *SimpleFlowEngine {
	return &SimpleFlowEngine{
		rules:       rules,
		modelClient: modelClient,
	}
}

// Execute 执行决策流
func (e *SimpleFlowEngine) Execute(ctx *DecisionContext) error {
	fmt.Println("[决策流] 开始执行")

	// 1. 执行规则集
	fmt.Println("[决策流] 步骤1: 执行规则集")
	fact := e.buildRuleFact(ctx.Input)
	ruleResults, err := rule.ExecuteRuleSet(e.rules, fact)
	if err != nil {
		return fmt.Errorf("execute rules: %w", err)
	}
	ctx.RuleResults = ruleResults

	// 2. 如果有规则拒绝，直接决策
	if ruleResults.AnyReject {
		ctx.Decision = "REJECT"
		ctx.DecisionCode = "REJECT_RULE"
		ctx.DecisionReason = ruleResults.FirstRejectReason
		fmt.Printf("[决策流] 规则拒绝: %s\n", ctx.DecisionReason)
		return nil
	}

	// 3. 调用模型
	fmt.Println("[决策流] 步骤2: 调用模型")
	modelReq := e.buildModelRequest(ctx.Input)
	modelResult, err := e.modelClient.Predict("M001", modelReq)
	if err != nil {
		return fmt.Errorf("call model: %w", err)
	}
	ctx.ModelResult = modelResult

	// 4. 根据模型结果决策
	fmt.Println("[决策流] 步骤3: 根据模型结果决策")
	e.makeDecision(ctx)

	fmt.Printf("[决策流] 完成: decision=%s, reason=%s\n", ctx.Decision, ctx.DecisionReason)
	return nil
}

// buildRuleFact 构建规则事实数据
func (e *SimpleFlowEngine) buildRuleFact(input map[string]interface{}) map[string]interface{} {
	fact := make(map[string]interface{})

	// 提取年龄
	if applicant, ok := input["applicant"].(map[string]interface{}); ok {
		if age, ok := applicant["age"].(float64); ok {
			fact["age"] = int(age)
		} else if age, ok := applicant["age"].(int); ok {
			fact["age"] = age
		}
		if income, ok := applicant["monthlyIncome"].(float64); ok {
			fact["monthlyIncome"] = int(income)
		} else if income, ok := applicant["monthlyIncome"].(int); ok {
			fact["monthlyIncome"] = income
		}
	}

	// 直接提取
	if age, ok := input["age"].(float64); ok {
		fact["age"] = int(age)
	} else if age, ok := input["age"].(int); ok {
		fact["age"] = age
	}
	if income, ok := input["monthlyIncome"].(float64); ok {
		fact["monthlyIncome"] = int(income)
	} else if income, ok := input["monthlyIncome"].(int); ok {
		fact["monthlyIncome"] = income
	}

	return fact
}

// buildModelRequest 构建模型请求
func (e *SimpleFlowEngine) buildModelRequest(input map[string]interface{}) *model.ModelRequest {
	req := &model.ModelRequest{
		Applicant:   make(map[string]interface{}),
		Application: make(map[string]interface{}),
	}

	if contractID, ok := input["contractId"].(string); ok {
		req.ContractID = contractID
	}

	if applicant, ok := input["applicant"].(map[string]interface{}); ok {
		req.Applicant = applicant
	}

	if application, ok := input["application"].(map[string]interface{}); ok {
		req.Application = application
	}

	return req
}

// makeDecision 根据模型结果进行决策
func (e *SimpleFlowEngine) makeDecision(ctx *DecisionContext) {
	riskScore := ctx.ModelResult.RiskScore

	switch {
	case riskScore >= 700:
		ctx.Decision = "APPROVE"
		ctx.DecisionCode = "APPROVE_MODEL"
		ctx.DecisionReason = "风险评分良好"
	case riskScore >= 600 && riskScore < 700:
		ctx.Decision = "MANUAL"
		ctx.DecisionCode = "MANUAL_REVIEW"
		ctx.DecisionReason = "需人工复核"
	default:
		ctx.Decision = "REJECT"
		ctx.DecisionCode = "REJECT_MODEL"
		ctx.DecisionReason = "风险评分不足"
	}
}
