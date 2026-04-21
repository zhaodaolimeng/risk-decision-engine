package flow

import (
	"fmt"

	"risk-decision-engine/internal/engine/datasource"
	"risk-decision-engine/internal/engine/model"
	"risk-decision-engine/internal/engine/rule"
)

// ComplexDecisionContext 复杂决策上下文
type ComplexDecisionContext struct {
	Input          map[string]interface{}
	DataSourceData map[string]*datasource.DataSourceResult
	RuleResults    *rule.RuleSetResult
	ModelResult    *model.ModelResult
	Decision       string
	DecisionReason string
	DecisionCode   string
}

// DataSourceConfig 数据源配置
type DataSourceConfig struct {
	ID       string
	Client   datasource.DataSourceService
	Required bool
}

// ComplexFlowEngine 复杂决策流引擎
type ComplexFlowEngine struct {
	dataSources []*DataSourceConfig
	rules       []*rule.SimpleRule
	modelClient model.ModelService
}

// NewComplexFlowEngine 创建复杂决策流引擎
func NewComplexFlowEngine(
	dataSources []*DataSourceConfig,
	rules []*rule.SimpleRule,
	modelClient model.ModelService,
) *ComplexFlowEngine {
	return &ComplexFlowEngine{
		dataSources: dataSources,
		rules:       rules,
		modelClient: modelClient,
	}
}

// Execute 执行决策流
func (e *ComplexFlowEngine) Execute(ctx *ComplexDecisionContext) error {
	fmt.Println("[复杂决策流] 开始执行")

	// 1. 预加载数据源
	fmt.Println("[复杂决策流] 步骤1: 预加载数据源")
	if err := e.loadDataSources(ctx); err != nil {
		return fmt.Errorf("load datasources: %w", err)
	}

	// 2. 执行规则集
	fmt.Println("[复杂决策流] 步骤2: 执行规则集")
	fact := e.buildRuleFact(ctx)
	ruleResults, err := rule.ExecuteRuleSet(e.rules, fact)
	if err != nil {
		return fmt.Errorf("execute rules: %w", err)
	}
	ctx.RuleResults = ruleResults

	// 3. 如果有规则拒绝，直接决策
	if ruleResults.AnyReject {
		ctx.Decision = "REJECT"
		ctx.DecisionCode = "REJECT_RULE"
		ctx.DecisionReason = ruleResults.FirstRejectReason
		fmt.Printf("[复杂决策流] 规则拒绝: %s\n", ctx.DecisionReason)
		return nil
	}

	// 4. 调用模型
	fmt.Println("[复杂决策流] 步骤3: 调用模型")
	modelReq := e.buildModelRequest(ctx)
	modelResult, err := e.modelClient.Predict("M001", modelReq)
	if err != nil {
		return fmt.Errorf("call model: %w", err)
	}
	ctx.ModelResult = modelResult

	// 5. 根据模型结果和数据源综合决策
	fmt.Println("[复杂决策流] 步骤4: 综合决策")
	e.makeComplexDecision(ctx)

	fmt.Printf("[复杂决策流] 完成: decision=%s, reason=%s\n", ctx.Decision, ctx.DecisionReason)
	return nil
}

// loadDataSources 加载所有数据源
func (e *ComplexFlowEngine) loadDataSources(ctx *ComplexDecisionContext) error {
	ctx.DataSourceData = make(map[string]*datasource.DataSourceResult)

	userID, _ := ctx.Input["userId"].(string)
	contractID, _ := ctx.Input["contractId"].(string)

	for _, dsConfig := range e.dataSources {
		req := &datasource.DataSourceRequest{
			ContractID: contractID,
			UserID:     userID,
			Params:     make(map[string]interface{}),
		}

		result, err := dsConfig.Client.Fetch(dsConfig.ID, req)
		if err != nil {
			if dsConfig.Required {
				return fmt.Errorf("datasource %s failed: %w", dsConfig.ID, err)
			}
			fmt.Printf("[复杂决策流] 可选数据源 %s 加载失败: %v\n", dsConfig.ID, err)
			continue
		}

		ctx.DataSourceData[dsConfig.ID] = result
	}

	return nil
}

// buildRuleFact 构建规则事实数据（合并输入和数据源）
func (e *ComplexFlowEngine) buildRuleFact(ctx *ComplexDecisionContext) map[string]interface{} {
	fact := make(map[string]interface{})

	// 1. 从Input复制
	if applicant, ok := ctx.Input["applicant"].(map[string]interface{}); ok {
		for k, v := range applicant {
			fact[k] = v
		}
	}

	// 2. 从数据源复制
	for dsID, dsResult := range ctx.DataSourceData {
		if dsResult.Data != nil {
			for k, v := range dsResult.Data {
				fact[k] = v
			}
		}
	}

	// 转换年龄格式
	if ageFloat, ok := fact["age"].(float64); ok {
		fact["age"] = int(ageFloat)
	}

	fmt.Printf("[复杂决策流] 规则事实数据 keys=%v\n", getFactKeys(fact))

	return fact
}

// buildModelRequest 构建模型请求
func (e *ComplexFlowEngine) buildModelRequest(ctx *ComplexDecisionContext) *model.ModelRequest {
	req := &model.ModelRequest{
		Applicant:   make(map[string]interface{}),
		Application: make(map[string]interface{}),
	}

	if contractID, ok := ctx.Input["contractId"].(string); ok {
		req.ContractID = contractID
	}

	if applicant, ok := ctx.Input["applicant"].(map[string]interface{}); ok {
		req.Applicant = applicant
	}

	if application, ok := ctx.Input["application"].(map[string]interface{}); ok {
		req.Application = application
	}

	// 合并数据源数据到上下文
	for dsID, dsResult := range ctx.DataSourceData {
		if dsResult.Data != nil {
			for k, v := range dsResult.Data {
				req.Applicant[k] = v
			}
		}
	}

	return req
}

// makeComplexDecision 综合决策
func (e *ComplexFlowEngine) makeComplexDecision(ctx *ComplexDecisionContext) {
	riskScore := ctx.ModelResult.RiskScore

	// 可调整阈值的逻辑
	var finalDecision string
	var finalReason string
	var finalCode string

	switch {
	case riskScore >= 700:
		finalDecision = "APPROVE"
		finalCode = "APPROVE_MODEL"
		finalReason = "风险评分良好"
	case riskScore >= 600 && riskScore < 700:
		// 检查多头数据
		if hasHighMulti(ctx) {
			finalDecision = "MANUAL"
			finalCode = "MANUAL_HIGH_MULTI"
			finalReason = "高多头借贷需人工复核"
		} else {
			finalDecision = "MANUAL"
			finalCode = "MANUAL_REVIEW"
			finalReason = "需人工复核"
		}
	default:
		finalDecision = "REJECT"
		finalCode = "REJECT_MODEL"
		finalReason = "风险评分不足"
	}

	ctx.Decision = finalDecision
	ctx.DecisionCode = finalCode
	ctx.DecisionReason = finalReason
}

// hasHighMulti 检查是否高多头
func hasHighMulti(ctx *ComplexDecisionContext) bool {
	// 从数据源检查多头数据
	for _, dsResult := range ctx.DataSourceData {
		if dsResult.Data != nil {
			if multiQuery, ok := dsResult.Data["multiQueryCount7d"].(float64); ok {
				if int(multiQuery) > 3 {
					return true
				}
			}
			if multiQuery, ok := dsResult.Data["multiQueryCount7d"].(int); ok {
				if multiQuery > 3 {
					return true
				}
			}
		}
	}
	return false
}

func getFactKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
