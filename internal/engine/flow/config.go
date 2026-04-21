package flow

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// FlowConfigWrapper 决策流配置包装（根节点）
type FlowConfigWrapper struct {
	Flow *FlowConfig `yaml:"flow"`
}

// FlowConfig 决策流配置
type FlowConfig struct {
	FlowID      string        `yaml:"flowId"`
	Version     string        `yaml:"version"`
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Status      string        `yaml:"status"`
	Preload     *PreloadConfig `yaml:"preload,omitempty"`
	Nodes       []*FlowNode   `yaml:"nodes"`
	Edges       []*FlowEdge   `yaml:"edges"`
	StartNodeID string        `yaml:"startNodeId"`
}

// PreloadConfig 预加载配置
type PreloadConfig struct {
	DatasourceIds []string `yaml:"datasourceIds"`
	Parallel      bool     `yaml:"parallel"`
	Timeout       string   `yaml:"timeout"`
}

// FlowNode 流节点
type FlowNode struct {
	NodeID           string                 `yaml:"nodeId"`
	Type             string                 `yaml:"type"` // START, RULE_SET, MODEL, DECISION, END
	RuleIds          []string               `yaml:"ruleIds,omitempty"`
	ModelID          string                 `yaml:"modelId,omitempty"`
	OutputToContext  string                 `yaml:"outputToContext,omitempty"`
	DecisionTable    *DecisionTableConfig  `yaml:"decisionTable,omitempty"`
}

// DecisionTableConfig 决策表配置
type DecisionTableConfig struct {
	Rules []*DecisionRuleConfig `yaml:"rules"`
}

// DecisionRuleConfig 决策规则配置
type DecisionRuleConfig struct {
	Condition string                 `yaml:"condition"`
	Result    string                 `yaml:"result"`
	Reason    string                 `yaml:"reason"`
	Extra     map[string]interface{} `yaml:"extra,omitempty"`
}

// FlowEdge 流边
type FlowEdge struct {
	From      string `yaml:"from"`
	To        string `yaml:"to"`
	Condition string `yaml:"condition,omitempty"`
}

// LoadFlowConfigFromFile 从YAML文件加载决策流配置
func LoadFlowConfigFromFile(filePath string) (*FlowConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var wrapper FlowConfigWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}

	if wrapper.Flow == nil {
		return nil, fmt.Errorf("no flow configuration found")
	}

	return wrapper.Flow, nil
}

// Validate 验证决策流配置
func (c *FlowConfig) Validate() error {
	if c.FlowID == "" {
		return fmt.Errorf("flowId is required")
	}
	if c.Status != "ACTIVE" && c.Status != "INACTIVE" {
		return fmt.Errorf("invalid status: %s", c.Status)
	}
	if len(c.Nodes) == 0 {
		return fmt.Errorf("no nodes defined")
	}
	if c.StartNodeID == "" {
		return fmt.Errorf("startNodeId is required")
	}

	// 验证开始节点存在
	startNodeFound := false
	nodeMap := make(map[string]*FlowNode)
	for _, node := range c.Nodes {
		nodeMap[node.NodeID] = node
		if node.NodeID == c.StartNodeID {
			startNodeFound = true
		}
	}
	if !startNodeFound {
		return fmt.Errorf("start node %s not found", c.StartNodeID)
	}

	// 验证边
	for _, edge := range c.Edges {
		if _, ok := nodeMap[edge.From]; !ok {
			return fmt.Errorf("edge from node %s not found", edge.From)
		}
		if _, ok := nodeMap[edge.To]; !ok {
			return fmt.Errorf("edge to node %s not found", edge.To)
		}
	}

	return nil
}

// IsActive 检查是否为激活状态
func (c *FlowConfig) IsActive() bool {
	return c.Status == "ACTIVE"
}

// GetNodeByID 根据ID获取节点
func (c *FlowConfig) GetNodeByID(nodeID string) *FlowNode {
	for _, node := range c.Nodes {
		if node.NodeID == nodeID {
			return node
		}
	}
	return nil
}

// GetOutgoingEdges 获取节点的出边
func (c *FlowConfig) GetOutgoingEdges(nodeID string) []*FlowEdge {
	var edges []*FlowEdge
	for _, edge := range c.Edges {
		if edge.From == nodeID {
			edges = append(edges, edge)
		}
	}
	return edges
}
