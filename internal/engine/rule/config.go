package rule

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// RuleConfig 规则配置文件根结构
type RuleConfig struct {
	Rules []*RuleDefinition `yaml:"rules"`
}

// RuleDefinition 规则定义
type RuleDefinition struct {
	RuleID      string          `yaml:"ruleId"`
	Version     string          `yaml:"version"`
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Type        string          `yaml:"type"`
	Priority    int             `yaml:"priority"`
	Status      string          `yaml:"status"`
	Condition   ConditionNode   `yaml:"condition"`
	Actions     ActionConfig    `yaml:"actions"`
}

// ConditionNode 条件节点
type ConditionNode struct {
	Operator    string           `yaml:"operator,omitempty"`   // AND/OR 或 比较操作符(>=, <=, ==等)
	Expressions []*ExpressionDef `yaml:"expressions,omitempty"`
	Field       string           `yaml:"field,omitempty"`
	Value       interface{}      `yaml:"value,omitempty"`
}

// ExpressionDef 表达式定义
type ExpressionDef struct {
	Field    string      `yaml:"field"`
	Operator string      `yaml:"operator"`
	Value    interface{} `yaml:"value"`
}

// ActionConfig 动作配置
type ActionConfig struct {
	True  *ActionDef `yaml:"true"`
	False *ActionDef `yaml:"false"`
}

// ActionDef 动作定义
type ActionDef struct {
	Result string `yaml:"result"`
	Reason string `yaml:"reason,omitempty"`
}

// LoadRuleConfigFromFile 从YAML文件加载规则配置
func LoadRuleConfigFromFile(filePath string) (*RuleConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var config RuleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}

	return &config, nil
}

// BuildExpression 从条件节点构建expr表达式
func (r *RuleDefinition) BuildExpression() (string, error) {
	return buildExprFromCondition(&r.Condition)
}

// buildExprFromCondition 递归构建表达式
func buildExprFromCondition(cond *ConditionNode) (string, error) {
	// 处理简单表达式（单个条件）
	if cond.Field != "" {
		return buildSingleExpression(cond.Field, cond.Operator, cond.Value)
	}

	// 处理复合表达式（AND/OR）
	if cond.Operator != "" && len(cond.Expressions) > 0 {
		var parts []string
		for _, expr := range cond.Expressions {
			part, err := buildSingleExpression(expr.Field, expr.Operator, expr.Value)
			if err != nil {
				return "", err
			}
			parts = append(parts, part)
		}

		op := strings.ToUpper(cond.Operator)
		if op == "AND" {
			return strings.Join(parts, " && "), nil
		} else if op == "OR" {
			return strings.Join(parts, " || "), nil
		}
		return "", fmt.Errorf("unknown operator: %s", cond.Operator)
	}

	return "", fmt.Errorf("invalid condition node")
}

// buildSingleExpression 构建单个表达式
func buildSingleExpression(field, operator string, value interface{}) (string, error) {
	// 转换字段名：将点号替换为下划线，或者保持原样（expr支持点号访问map）
	// expr支持: DS001.isBlacklist 或者 input.applicant.age
	exprField := field

	// 转换操作符
	var exprOp string
	switch operator {
	case "==":
		exprOp = "=="
	case "!=":
		exprOp = "!="
	case ">":
		exprOp = ">"
	case ">=":
		exprOp = ">="
	case "<":
		exprOp = "<"
	case "<=":
		exprOp = "<="
	default:
		return "", fmt.Errorf("unknown operator: %s", operator)
	}

	// 格式化值
	var valueStr string
	switch v := value.(type) {
	case string:
		valueStr = fmt.Sprintf("%q", v)
	case int, int32, int64, float32, float64:
		valueStr = fmt.Sprintf("%v", v)
	case bool:
		valueStr = fmt.Sprintf("%t", v)
	default:
		return "", fmt.Errorf("unsupported value type: %T", v)
	}

	return fmt.Sprintf("%s %s %s", exprField, exprOp, valueStr), nil
}

// ToSimpleRule 将规则定义转换为SimpleRule
func (r *RuleDefinition) ToSimpleRule() (*SimpleRule, error) {
	expression, err := r.BuildExpression()
	if err != nil {
		return nil, fmt.Errorf("build expression for %s: %w", r.RuleID, err)
	}

	rule := &SimpleRule{
		ID:         r.RuleID,
		Name:       r.Name,
		Expression: expression,
	}

	if r.Actions.True != nil {
		rule.PassAction = &Action{
			Result: r.Actions.True.Result,
		}
	}

	if r.Actions.False != nil {
		rule.FailAction = &Action{
			Result: r.Actions.False.Result,
			Reason: r.Actions.False.Reason,
		}
	}

	return rule, nil
}

// LoadRulesFromConfig 从配置加载所有规则
func LoadRulesFromConfig(config *RuleConfig) ([]*SimpleRule, error) {
	var rules []*SimpleRule

	for _, ruleDef := range config.Rules {
		// 只加载ACTIVE状态的规则
		if ruleDef.Status != "ACTIVE" {
			continue
		}

		rule, err := ruleDef.ToSimpleRule()
		if err != nil {
			return nil, fmt.Errorf("convert rule %s: %w", ruleDef.RuleID, err)
		}
		rules = append(rules, rule)
	}

	// 按优先级排序（数字越大优先级越高）
	for i := range rules {
		for j := i + 1; j < len(rules); j++ {
			// 从RuleDefinition找回优先级信息
			var pi, pj int
			for _, rd := range config.Rules {
				if rd.RuleID == rules[i].ID {
					pi = rd.Priority
				}
				if rd.RuleID == rules[j].ID {
					pj = rd.Priority
				}
			}
			if pj > pi {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}

	return rules, nil
}

// LoadRulesFromFile 从文件加载规则
func LoadRulesFromFile(filePath string) ([]*SimpleRule, error) {
	config, err := LoadRuleConfigFromFile(filePath)
	if err != nil {
		return nil, err
	}
	return LoadRulesFromConfig(config)
}
