package rule

import (
	"encoding/json"
	"fmt"
	"sync"

	"risk-decision-engine/internal/engine/rule"
	"risk-decision-engine/pkg/logger"
)

// Service 规则服务
type Service struct {
	rules map[string]*rule.Rule
	mu    sync.RWMutex
}

// NewService 创建规则服务
func NewService() *Service {
	return &Service{
		rules: make(map[string]*rule.Rule),
	}
}

// LoadRule 加载规则
func (s *Service) LoadRule(data []byte) (*rule.Rule, error) {
	r, err := rule.LoadRuleFromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("load rule: %w", err)
	}

	s.mu.Lock()
	s.rules[r.RuleID] = r
	s.mu.Unlock()

	logger.Infof("Rule loaded: %s (v%s)", r.Name, r.Version)
	return r, nil
}

// GetRule 获取规则
func (s *Service) GetRule(ruleID string) (*rule.Rule, bool) {
	s.mu.RLock()
	r, ok := s.rules[ruleID]
	s.mu.RUnlock()
	return r, ok
}

// GetAllRules 获取所有规则
func (s *Service) GetAllRules() map[string]*rule.Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*rule.Rule)
	for k, v := range s.rules {
		result[k] = v
	}
	return result
}

// LoadAgeRule 加载测试用的年龄规则
func (s *Service) LoadAgeRule() (*rule.Rule, error) {
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

	return s.LoadRule(ruleJSON)
}

// ExecuteSimpleDecision 执行简单决策（单规则）
func (s *Service) ExecuteSimpleDecision(ruleID string, fact map[string]interface{}) (string, string, error) {
	r, ok := s.GetRule(ruleID)
	if !ok {
		return "", "", fmt.Errorf("rule not found: %s", ruleID)
	}

	result, err := r.Execute(fact)
	if err != nil {
		return "", "", fmt.Errorf("execute rule: %w", err)
	}

	if result.Action == nil {
		return "PASS", "", nil
	}

	return result.Action.Result, result.Action.Reason, nil
}

// DecisionRequest 决策请求
type DecisionRequest struct {
	RequestID  string                 `json:"requestId"`
	BusinessID string                 `json:"businessId"`
	ContractID string                 `json:"contractId,omitempty"`
	Data       map[string]interface{} `json:"data"`
}

// DecisionResponse 决策响应
type DecisionResponse struct {
	DecisionID     string                 `json:"decisionId"`
	BusinessID     string                 `json:"businessId"`
	Decision       string                 `json:"decision"`
	DecisionCode   string                 `json:"decisionCode,omitempty"`
	DecisionReason string                 `json:"decisionReason,omitempty"`
	RuleResults    []*rule.RuleResult     `json:"ruleResults,omitempty"`
	ModelResults   map[string]interface{} `json:"modelResults,omitempty"`
}
