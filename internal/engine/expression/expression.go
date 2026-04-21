package expression

import (
	"fmt"

	"github.com/expr-lang/expr"
)

// Expression 表达式接口
type Expression interface {
	Evaluate(env map[string]interface{}) (interface{}, error)
}

// SimpleExpression 简单表达式
type SimpleExpression struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// CompoundExpression 复合表达式
type CompoundExpression struct {
	Operator     string           `json:"operator"` // AND, OR
	Expressions   []Expression     `json:"expressions"`
}

// ExprExpression 使用 expr 库的表达式
type ExprExpression struct {
	Expr string
}

// NewExprExpression 创建表达式
func NewExprExpression(exprStr string) *ExprExpression {
	return &ExprExpression{Expr: exprStr}
}

// Evaluate 执行表达式
func (e *ExprExpression) Evaluate(env map[string]interface{}) (interface{}, error) {
	output, err := expr.Eval(e.Expr, env)
	if err != nil {
		return nil, fmt.Errorf("eval expression '%s': %w", e.Expr, err)
	}
	return output, nil
}

// EvaluateBoolean 执行布尔表达式
func EvaluateBoolean(exprStr string, env map[string]interface{}) (bool, error) {
	output, err := expr.Eval(exprStr, env)
	if err != nil {
		return false, fmt.Errorf("eval boolean expression '%s': %w", exprStr, err)
	}

	boolVal, ok := output.(bool)
	if !ok {
		return false, fmt.Errorf("expression result is not boolean: %v", output)
	}

	return boolVal, nil
}

// BuildConditionExpr 从简单条件构建表达式字符串
func BuildConditionExpr(field string, operator string, value interface{}) string {
	switch operator {
	case "==":
		return fmt.Sprintf("%s == %v", field, formatValue(value))
	case "!=":
		return fmt.Sprintf("%s != %v", field, formatValue(value))
	case ">":
		return fmt.Sprintf("%s > %v", field, formatValue(value))
	case ">=":
		return fmt.Sprintf("%s >= %v", field, formatValue(value))
	case "<":
		return fmt.Sprintf("%s < %v", field, formatValue(value))
	case "<=":
		return fmt.Sprintf("%s <= %v", field, formatValue(value))
	case "IN":
		return fmt.Sprintf("%s in %v", field, value)
	case "NOT_IN":
		return fmt.Sprintf("%s not in %v", field, value)
	case "CONTAINS":
		return fmt.Sprintf(`contains(%s, "%v")`, field, value)
	default:
		return fmt.Sprintf("%s %s %v", field, operator, formatValue(value))
	}
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
