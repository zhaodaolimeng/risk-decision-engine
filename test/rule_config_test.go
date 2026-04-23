package test

import (
	"path/filepath"
	"testing"

	"risk-decision-engine/internal/engine/rule"

	"github.com/stretchr/testify/assert"
)

// TestLoadRuleConfigFromFile 测试从文件加载规则配置
func TestLoadRuleConfigFromFile(t *testing.T) {
	configPath := filepath.Join("cases", "simple", "01-age-rule", "config", "rule.yaml")

	t.Run("LoadValidConfig", func(t *testing.T) {
		config, err := rule.LoadRuleConfigFromFile(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Len(t, config.Rules, 1)
		assert.Equal(t, "R001", config.Rules[0].RuleID)
		assert.Equal(t, "ACTIVE", config.Rules[0].Status)
	})

	t.Run("LoadInvalidPath", func(t *testing.T) {
		config, err := rule.LoadRuleConfigFromFile("invalid/path.yaml")
		assert.Error(t, err)
		assert.Nil(t, config)
	})
}

// TestLoadRulesFromFile 测试从文件加载规则
func TestLoadRulesFromFile(t *testing.T) {
	configPath := filepath.Join("cases", "simple", "01-age-rule", "config", "rule.yaml")

	t.Run("LoadRules", func(t *testing.T) {
		rules, err := rule.LoadRulesFromFile(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, rules)
		assert.GreaterOrEqual(t, len(rules), 1)
	})
}

// TestBuildExpression 测试构建表达式
func TestBuildExpression(t *testing.T) {
	configPath := filepath.Join("cases", "simple", "01-age-rule", "config", "rule.yaml")
	config, err := rule.LoadRuleConfigFromFile(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, config)

	if len(config.Rules) > 0 {
		t.Run("BuildExpression", func(t *testing.T) {
			expr, err := config.Rules[0].BuildExpression()
			assert.NoError(t, err)
			assert.NotEmpty(t, expr)
			t.Logf("Expression: %s", expr)
		})
	}
}
