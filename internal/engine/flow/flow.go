package flow

import (
	"encoding/json"
	"fmt"
	"time"

	"risk-decision-engine/internal/engine/rule"
	"risk-decision-engine/pkg/logger"
)

// NodeType 节点类型
type NodeType string

const (
	NodeTypeStart    NodeType = "START"
	NodeTypeEnd      NodeType = "END"
	NodeTypeRuleSet  NodeType = "RULE_SET"
	NodeTypeDecision NodeType = "DECISION"
)

// Decision 决策结果
const (
	DecisionApprove = "APPROVE"
	DecisionReject  = "REJECT"
	DecisionManual  = "MANUAL"
)

// Flow 决策流定义
type Flow struct {
	FlowID      string                 `json:"flowId"`
	Version     string                 `json:"version"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Nodes       []Node                 `json:"nodes"`
	Edges       []Edge                 `json:"edges"`
	StartNodeID string                 `json:"startNodeId"`

	// 内部字段
	nodesMap map[string]Node
	edgesMap map[string][]Edge
}

// Node 节点
type Node struct {
	NodeID   string                 `json:"nodeId"`
	Type     NodeType               `json:"type"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

// Edge 连线
type Edge struct {
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Condition map[string]interface{} `json:"condition,omitempty"`
}

// NodeResult 节点执行结果
type NodeResult struct {
	NodeID      string        `json:"nodeId"`
	NodeType    NodeType      `json:"nodeType"`
	Output      interface{}   `json:"output,omitempty"`
	RuleResults []*rule.RuleResult `json:"ruleResults,omitempty"`
	ExecuteTime time.Duration `json:"executeTime,omitempty"`
}

// FlowResult 决策流结果
type FlowResult struct {
	FlowID      string        `json:"flowId"`
	FlowName    string        `json:"flowName"`
	Decision    string        `json:"decision"`
	DecisionCode string        `json:"decisionCode,omitempty"`
	Reason      string        `json:"reason"`
	NodeResults []*NodeResult `json:"nodeResults,omitempty"`
	ExecuteTime time.Duration `json:"executeTime,omitempty"`
}

// LoadFlowFromJSON 从JSON加载决策流
func LoadFlowFromJSON(data []byte) (*Flow, error) {
	var flow Flow
	if err := json.Unmarshal(data, &flow); err != nil {
		return nil, fmt.Errorf("unmarshal flow: %w", err)
	}

	flow.InitMaps()
	return &flow, nil
}

// InitMaps 初始化节点和边的映射
func (f *Flow) InitMaps() {
	f.nodesMap = make(map[string]Node)
	for _, node := range f.Nodes {
		f.nodesMap[node.NodeID] = node
	}

	f.edgesMap = make(map[string][]Edge)
	for _, edge := range f.Edges {
		f.edgesMap[edge.From] = append(f.edgesMap[edge.From], edge)
	}
}

// Execute 执行决策流
func (f *Flow) Execute(fact map[string]interface{}, rules map[string]*rule.Rule) (*FlowResult, error) {
	start := time.Now()
	result := &FlowResult{
		FlowID:   f.FlowID,
		FlowName: f.Name,
	}

	currentNodeID := f.StartNodeID
	var nodeResults []*NodeResult

	// 上下文数据
	ctxData := make(map[string]interface{})
	for k, v := range fact {
		ctxData[k] = v
	}

	for currentNodeID != "" {
		node, ok := f.nodesMap[currentNodeID]
		if !ok {
			return nil, fmt.Errorf("node not found: %s", currentNodeID)
		}

		logger.Debugf("Executing node: %s (%s)", node.NodeID, node.Type)

		nodeResult, nextNodeID, err := f.executeNode(&node, ctxData, rules)
		if err != nil {
			return nil, fmt.Errorf("execute node %s: %w", node.NodeID, err)
		}

		if nodeResult != nil {
			nodeResults = append(nodeResults, nodeResult)
		}

		// 如果是结束节点，退出
		if node.Type == NodeTypeEnd {
			break
		}

		currentNodeID = nextNodeID
	}

	result.NodeResults = nodeResults
	result.ExecuteTime = time.Since(start)

	// 从决策节点获取结果
	for _, nr := range nodeResults {
		if nr.NodeType == NodeTypeDecision {
			if output, ok := nr.Output.(map[string]interface{}); ok {
				if decision, ok := output["decision"]; ok {
					result.Decision = fmt.Sprintf("%v", decision)
				}
				if reason, ok := output["reason"]; ok {
					result.Reason = fmt.Sprintf("%v", reason)
				}
			}
		}
	}

	return result, nil
}

func (f *Flow) executeNode(node *Node, ctxData map[string]interface{}, rules map[string]*rule.Rule) (*NodeResult, string, error) {
	start := time.Now()
	result := &NodeResult{
		NodeID:   node.NodeID,
		NodeType: node.Type,
	}

	var nextNodeID string
	var err error

	switch node.Type {
	case NodeTypeStart:
		nextNodeID = f.getNextNode(node.NodeID, ctxData)

	case NodeTypeRuleSet:
		result.RuleResults, err = f.executeRuleSet(node, ctxData, rules)
		if err != nil {
			return nil, "", err
		}
		// 更新上下文
		for _, rr := range result.RuleResults {
			ctxData[rr.RuleID] = rr.Pass
			if rr.Action != nil {
				ctxData[rr.RuleID+"_result"] = rr.Action.Result
				ctxData[rr.RuleID+"_reason"] = rr.Action.Reason
			}
		}
		// 检查是否有拒绝的规则
		anyReject := false
		firstRejectReason := ""
		for _, rr := range result.RuleResults {
			if rr.Action != nil && rr.Action.Result == "REJECT" {
				anyReject = true
				firstRejectReason = rr.Action.Reason
				break
			}
		}
		ctxData["anyRuleReject"] = anyReject
		ctxData["firstRejectReason"] = firstRejectReason
		ctxData["allRulesPass"] = !anyReject

		nextNodeID = f.getNextNode(node.NodeID, ctxData)

	case NodeTypeDecision:
		output, err := f.executeDecision(node, ctxData)
		if err != nil {
			return nil, "", err
		}
		result.Output = output
		nextNodeID = f.getNextNode(node.NodeID, ctxData)

	case NodeTypeEnd:
		// 结束
	}

	result.ExecuteTime = time.Since(start)
	return result, nextNodeID, nil
}

func (f *Flow) executeRuleSet(node *Node, ctxData map[string]interface{}, rules map[string]*rule.Rule) ([]*rule.RuleResult, error) {
	var results []*rule.RuleResult

	ruleIDs, ok := node.Config["ruleIds"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("ruleIds not found or invalid format")
	}

	for _, rid := range ruleIDs {
		ruleID := fmt.Sprintf("%v", rid)
		r, ok := rules[ruleID]
		if !ok {
			logger.Warnf("Rule not found: %s", ruleID)
			continue
		}

		result, err := r.Execute(ctxData)
		if err != nil {
			return nil, fmt.Errorf("execute rule %s: %w", ruleID, err)
		}
		results = append(results, result)
	}

	return results, nil
}

func (f *Flow) executeDecision(node *Node, ctxData map[string]interface{}) (map[string]interface{}, error) {
	output := make(map[string]interface{})

	decisionTable, ok := node.Config["decisionTable"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("decisionTable not found")
	}

	rules, ok := decisionTable["rules"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("decisionTable.rules not found")
	}

	for _, r := range rules {
		ruleMap, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		condition, ok := ruleMap["condition"].(string)
		if !ok {
			continue
		}

		// 简单的条件判断
		pass := evaluateSimpleCondition(condition, ctxData)
		if pass {
			if result, ok := ruleMap["result"]; ok {
				output["decision"] = result
			}
			if reason, ok := ruleMap["reason"]; ok {
				output["reason"] = reason
			}
			break
		}
	}

	return output, nil
}

func evaluateSimpleCondition(condition string, ctxData map[string]interface{}) bool {
	// 简单的条件判断，支持:
	// anyRuleReject == true
	// allRulesPass == true
	if condition == "anyRuleReject == true" {
		if v, ok := ctxData["anyRuleReject"]; ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
		return false
	}
	if condition == "allRulesPass == true" {
		if v, ok := ctxData["allRulesPass"]; ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
		return true
	}
	// 默认返回 false
	return false
}

func (f *Flow) getNextNode(fromNodeID string, ctxData map[string]interface{}) string {
	edges := f.edgesMap[fromNodeID]
	if len(edges) == 0 {
		return ""
	}

	// 返回第一个边的目标
	return edges[0].To
}
