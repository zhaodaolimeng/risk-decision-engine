package sandbox

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"time"

	"risk-decision-engine/internal/engine/rule"
)

// DiffType 差异类型
type DiffType string

const (
	DiffTypeAdded    DiffType = "ADDED"
	DiffTypeRemoved  DiffType = "REMOVED"
	DiffTypeModified DiffType = "MODIFIED"
	DiffTypeUnchanged DiffType = "UNCHANGED"
)

// FieldDiff 字段差异
type FieldDiff struct {
	FieldName string      `json:"fieldName"`
	OldValue  interface{} `json:"oldValue,omitempty"`
	NewValue  interface{} `json:"newValue,omitempty"`
	DiffType  DiffType    `json:"diffType"`
}

// RuleDiff 规则差异
type RuleDiff struct {
	RuleID      string      `json:"ruleId"`
	RuleName    string      `json:"ruleName,omitempty"`
	DiffType    DiffType    `json:"diffType"`
	FieldDiffs  []*FieldDiff `json:"fieldDiffs,omitempty"`
	OldHash     string      `json:"oldHash,omitempty"`
	NewHash     string      `json:"newHash,omitempty"`
}

// ConfigDiffReport 配置比对报告
type ConfigDiffReport struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	OldConfigPath string      `json:"oldConfigPath,omitempty"`
	NewConfigPath string      `json:"newConfigPath,omitempty"`
	TotalRules   int         `json:"totalRules"`
	AddedCount   int         `json:"addedCount"`
	RemovedCount int         `json:"removedCount"`
	ModifiedCount int         `json:"modifiedCount"`
	UnchangedCount int        `json:"unchangedCount"`
	RuleDiffs    []*RuleDiff `json:"ruleDiffs"`
	HasBreakingChanges bool   `json:"hasBreakingChanges"`
	BreakingChanges []string `json:"breakingChanges,omitempty"`
}

// RuleComparator 规则比对器
type RuleComparator struct {
}

// NewRuleComparator 创建规则比对器
func NewRuleComparator() *RuleComparator {
	return &RuleComparator{}
}

// CompareRuleConfigs 比对两个规则配置
func (c *RuleComparator) CompareRuleConfigs(
	name string,
	oldConfig *rule.RuleConfig,
	newConfig *rule.RuleConfig,
) (*ConfigDiffReport, error) {

	report := &ConfigDiffReport{
		ID:       "diff_" + generateID(),
		Name:     name,
		RuleDiffs: make([]*RuleDiff, 0),
	}

	// 构建规则映射
	oldRules := make(map[string]*rule.RuleDefinition)
	for _, r := range oldConfig.Rules {
		oldRules[r.RuleID] = r
	}

	newRules := make(map[string]*rule.RuleDefinition)
	for _, r := range newConfig.Rules {
		newRules[r.RuleID] = r
	}

	// 检查新增和修改的规则
	for ruleID, newRule := range newRules {
		oldRule, exists := oldRules[ruleID]
		if !exists {
			// 新增规则
			diff := &RuleDiff{
				RuleID:   ruleID,
				RuleName: newRule.Name,
				DiffType: DiffTypeAdded,
				NewHash:  c.hashRule(newRule),
			}
			report.RuleDiffs = append(report.RuleDiffs, diff)
			report.AddedCount++
		} else {
			// 检查是否修改
			fieldDiffs := c.compareRules(oldRule, newRule)
			if len(fieldDiffs) > 0 {
				diff := &RuleDiff{
					RuleID:   ruleID,
					RuleName: newRule.Name,
					DiffType: DiffTypeModified,
					FieldDiffs: fieldDiffs,
					OldHash:  c.hashRule(oldRule),
					NewHash:  c.hashRule(newRule),
				}
				report.RuleDiffs = append(report.RuleDiffs, diff)
				report.ModifiedCount++
			} else {
				// 未修改
				diff := &RuleDiff{
					RuleID:   ruleID,
					RuleName: newRule.Name,
					DiffType: DiffTypeUnchanged,
					OldHash:  c.hashRule(oldRule),
					NewHash:  c.hashRule(newRule),
				}
				report.RuleDiffs = append(report.RuleDiffs, diff)
				report.UnchangedCount++
			}
		}
	}

	// 检查删除的规则
	for ruleID, oldRule := range oldRules {
		if _, exists := newRules[ruleID]; !exists {
			diff := &RuleDiff{
				RuleID:   ruleID,
				RuleName: oldRule.Name,
				DiffType: DiffTypeRemoved,
				OldHash:  c.hashRule(oldRule),
			}
			report.RuleDiffs = append(report.RuleDiffs, diff)
			report.RemovedCount++
		}
	}

	report.TotalRules = len(oldConfig.Rules)
	report.HasBreakingChanges = c.checkBreakingChanges(report)
	if report.HasBreakingChanges {
		report.BreakingChanges = c.identifyBreakingChanges(report)
	}

	return report, nil
}

// compareRules 比对两个规则
func (c *RuleComparator) compareRules(oldRule, newRule *rule.RuleDefinition) []*FieldDiff {
	var diffs []*FieldDiff

	// 比对基本字段
	if oldRule.Name != newRule.Name {
		diffs = append(diffs, &FieldDiff{
			FieldName: "name",
			OldValue:  oldRule.Name,
			NewValue:  newRule.Name,
			DiffType:  DiffTypeModified,
		})
	}

	if oldRule.Description != newRule.Description {
		diffs = append(diffs, &FieldDiff{
			FieldName: "description",
			OldValue:  oldRule.Description,
			NewValue:  newRule.Description,
			DiffType:  DiffTypeModified,
		})
	}

	if oldRule.Status != newRule.Status {
		diffs = append(diffs, &FieldDiff{
			FieldName: "status",
			OldValue:  oldRule.Status,
			NewValue:  newRule.Status,
			DiffType:  DiffTypeModified,
		})
	}

	if oldRule.Priority != newRule.Priority {
		diffs = append(diffs, &FieldDiff{
			FieldName: "priority",
			OldValue:  oldRule.Priority,
			NewValue:  newRule.Priority,
			DiffType:  DiffTypeModified,
		})
	}

	// 比对条件
	oldExpr, _ := oldRule.BuildExpression()
	newExpr, _ := newRule.BuildExpression()
	if oldExpr != newExpr {
		diffs = append(diffs, &FieldDiff{
			FieldName: "expression",
			OldValue:  oldExpr,
			NewValue:  newExpr,
			DiffType:  DiffTypeModified,
		})
	}

	// 比对动作
	if !reflect.DeepEqual(oldRule.Actions, newRule.Actions) {
		diffs = append(diffs, &FieldDiff{
			FieldName: "actions",
			OldValue:  oldRule.Actions,
			NewValue:  newRule.Actions,
			DiffType:  DiffTypeModified,
		})
	}

	return diffs
}

// hashRule 计算规则的哈希值
func (c *RuleComparator) hashRule(r *rule.RuleDefinition) string {
	expr, _ := r.BuildExpression()
	data := fmt.Sprintf("%s|%s|%s|%d|%s|%v",
		r.RuleID, r.Name, r.Status, r.Priority, expr, r.Actions)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// checkBreakingChanges 检查是否有破坏性变更
func (c *RuleComparator) checkBreakingChanges(report *ConfigDiffReport) bool {
	for _, diff := range report.RuleDiffs {
		if diff.DiffType == DiffTypeRemoved {
			return true
		}
		if diff.DiffType == DiffTypeModified {
			for _, fd := range diff.FieldDiffs {
				if fd.FieldName == "expression" || fd.FieldName == "actions" {
					return true
				}
			}
		}
	}
	return false
}

// identifyBreakingChanges 识别破坏性变更
func (c *RuleComparator) identifyBreakingChanges(report *ConfigDiffReport) []string {
	var changes []string
	for _, diff := range report.RuleDiffs {
		if diff.DiffType == DiffTypeRemoved {
			changes = append(changes, fmt.Sprintf("规则被删除: %s (%s)", diff.RuleID, diff.RuleName))
		}
		if diff.DiffType == DiffTypeModified {
			for _, fd := range diff.FieldDiffs {
				if fd.FieldName == "expression" {
					changes = append(changes, fmt.Sprintf("规则表达式变更: %s (%s)", diff.RuleID, diff.RuleName))
				}
				if fd.FieldName == "actions" {
					changes = append(changes, fmt.Sprintf("规则动作变更: %s (%s)", diff.RuleID, diff.RuleName))
				}
			}
		}
	}
	return changes
}

// GenerateSummary 生成比对摘要
func (r *ConfigDiffReport) GenerateSummary() string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("规则配置比对报告: %s\n", r.Name))
	summary.WriteString(fmt.Sprintf("总规则数: %d\n", r.TotalRules))
	summary.WriteString(fmt.Sprintf("新增: %d, 删除: %d, 修改: %d, 未变: %d\n",
		r.AddedCount, r.RemovedCount, r.ModifiedCount, r.UnchangedCount))

	if r.HasBreakingChanges {
		summary.WriteString("\n⚠️  检测到破坏性变更:\n")
		for _, change := range r.BreakingChanges {
			summary.WriteString(fmt.Sprintf("  - %s\n", change))
		}
	}

	return summary.String()
}

// GetModifiedRules 获取修改的规则
func (r *ConfigDiffReport) GetModifiedRules() []*RuleDiff {
	var modified []*RuleDiff
	for _, diff := range r.RuleDiffs {
		if diff.DiffType == DiffTypeModified {
			modified = append(modified, diff)
		}
	}
	return modified
}

// GetAddedRules 获取新增的规则
func (r *ConfigDiffReport) GetAddedRules() []*RuleDiff {
	var added []*RuleDiff
	for _, diff := range r.RuleDiffs {
		if diff.DiffType == DiffTypeAdded {
			added = append(added, diff)
		}
	}
	return added
}

// GetRemovedRules 获取删除的规则
func (r *ConfigDiffReport) GetRemovedRules() []*RuleDiff {
	var removed []*RuleDiff
	for _, diff := range r.RuleDiffs {
		if diff.DiffType == DiffTypeRemoved {
			removed = append(removed, diff)
		}
	}
	return removed
}

func generateID() string {
	// 简单的ID生成
	hash := md5.Sum([]byte(time.Now().String()))
	return "cmp_" + hex.EncodeToString(hash[:])[:8]
}
