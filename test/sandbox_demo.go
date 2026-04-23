package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080"

func main() {
	fmt.Println("========================================")
	fmt.Println("   沙盒功能演示")
	fmt.Println("========================================")
	fmt.Println()

	// 检查服务是否启动
	if !healthCheck() {
		fmt.Println("请先启动沙盒服务: go run cmd/server/sandbox_server.go")
		return
	}

	// 演示1: 开始记录会话
	fmt.Println("步骤1: 开始流量记录")
	sessionID := startRecording("测试会话-1")
	if sessionID == "" {
		fmt.Println("开始记录失败")
		return
	}
	fmt.Printf("✓ 记录会话已开始: %s\n\n", sessionID)

	// 演示2: 发送测试请求（会被记录）
	fmt.Println("步骤2: 发送决策请求（记录中）")
	testCases := []int{25, 18, 30, 65, 20}
	for i, age := range testCases {
		time.Sleep(100 * time.Millisecond)
		reqID := fmt.Sprintf("test-%03d", i+1)
		result := executeDecision(reqID, age)
		fmt.Printf("  [%s] 年龄=%d → %s\n", reqID, age, result)
	}
	fmt.Println()

	// 演示3: 停止记录
	fmt.Println("步骤3: 停止流量记录")
	stopRecording()
	fmt.Println("✓ 记录会话已停止\n")

	// 演示4: 查看记录会话详情
	fmt.Println("步骤4: 查看记录会话")
	records := getRecords(sessionID)
	fmt.Printf("✓ 共记录 %d 条请求\n\n", len(records))

	// 演示5: 规则配置比对
	fmt.Println("步骤5: 比对规则配置差异")
	diffReport := compareConfigs(
		"配置v1与v2比对",
		"test/cases/simple/01-age-rule/config/rule.yaml",
		"test/cases/simple/01-age-rule/config/rule_v2.yaml",
	)
	if diffReport != nil {
		fmt.Printf("✓ 比对完成: 新增=%d, 删除=%d, 修改=%d\n",
			diffReport.AddedCount, diffReport.RemovedCount, diffReport.ModifiedCount)
		if diffReport.HasBreakingChanges {
			fmt.Println("⚠ 检测到破坏性变更!")
		}
	}
	fmt.Println()

	// 演示6: 开始回放（使用v2配置）
	fmt.Println("步骤6: 开始流量回放（使用v2配置）")
	replayID := startReplayWithOptions(
		"回放测试-v2",
		sessionID,
		"test/cases/simple/01-age-rule/config/rule_v2.yaml",
	)
	if replayID == "" {
		fmt.Println("开始回放失败")
		return
	}
	fmt.Printf("✓ 回放已开始: %s\n\n", replayID)

	// 等待回放完成
	fmt.Println("步骤7: 等待回放完成...")
	time.Sleep(2 * time.Second)

	// 演示7: 获取回放报告
	fmt.Println("步骤8: 获取回放报告")
	report := getReplayReport(replayID)
	if report != nil {
		fmt.Printf("✓ 回放完成: 总数=%d, 匹配=%d, 不匹配=%d, 匹配率=%.2f%%\n",
			report.TotalRequests, report.MatchedCount, report.MismatchedCount, report.MatchRate)
		if len(report.Mismatches) > 0 {
			fmt.Println("\n  不匹配的请求:")
			for _, m := range report.Mismatches {
				fmt.Printf("    - [%s] 原决策=%s/%s → 新决策=%s/%s\n",
					m.RequestID, m.OriginalDecision, m.OriginalCode, m.NewDecision, m.NewCode)
			}
		}
	}
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("   演示完成!")
	fmt.Println("========================================")
}

// API 辅助函数

func healthCheck() bool {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func startRecording(name string) string {
	reqBody := map[string]string{"name": name}
	data, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL+"/api/v1/sandbox/record/start", "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Println("请求失败:", err)
		return ""
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["code"] != "0000" {
		fmt.Println("错误:", result["message"])
		return ""
	}

	session := result["data"].(map[string]interface{})
	return session["id"].(string)
}

func stopRecording() {
	resp, err := http.Post(baseURL+"/api/v1/sandbox/record/stop", "application/json", nil)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()
}

func executeDecision(requestID string, age int) string {
	reqBody := map[string]interface{}{
		"requestId": requestID,
		"data": map[string]interface{}{
			"age": age,
		},
	}
	data, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL+"/api/v1/decision/execute", "application/json", bytes.NewReader(data))
	if err != nil {
		return "请求失败"
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["code"] != "0000" {
		return "错误:" + result["message"].(string)
	}

	dataMap := result["data"].(map[string]interface{})
	return dataMap["decision"].(string)
}

func getRecords(sessionID string) []interface{} {
	resp, err := http.Get(baseURL + "/api/v1/sandbox/record/sessions/" + sessionID + "/records")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["code"] != "0000" {
		return nil
	}

	data := result["data"].(map[string]interface{})
	return data["records"].([]interface{})
}

type DiffReport struct {
	AddedCount      int    `json:"addedCount"`
	RemovedCount    int    `json:"removedCount"`
	ModifiedCount   int    `json:"modifiedCount"`
	HasBreakingChanges bool `json:"hasBreakingChanges"`
}

func compareConfigs(name, oldPath, newPath string) *DiffReport {
	reqBody := map[string]string{
		"name":         name,
		"oldConfigPath": oldPath,
		"newConfigPath": newPath,
	}
	data, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL+"/api/v1/sandbox/diff/rules", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["code"] != "0000" {
		return nil
	}

	dataMap := result["data"].(map[string]interface{})
	return &DiffReport{
		AddedCount:         int(dataMap["addedCount"].(float64)),
		RemovedCount:       int(dataMap["removedCount"].(float64)),
		ModifiedCount:      int(dataMap["modifiedCount"].(float64)),
		HasBreakingChanges: dataMap["hasBreakingChanges"].(bool),
	}
}

func startReplayWithOptions(name, recordingID, configPath string) string {
	reqBody := map[string]interface{}{
		"name":           name,
		"recordingId":    recordingID,
		"ruleConfigPath": configPath,
	}
	data, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL+"/api/v1/sandbox/replay/start/options", "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Println("请求失败:", err)
		return ""
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["code"] != "0000" {
		fmt.Println("错误:", result["message"])
		return ""
	}

	session := result["data"].(map[string]interface{})
	return session["id"].(string)
}

type ReplayReport struct {
	TotalRequests   int            `json:"totalRequests"`
	MatchedCount    int            `json:"matchedCount"`
	MismatchedCount int            `json:"mismatchedCount"`
	MatchRate       float64        `json:"matchRate"`
	Mismatches      []interface{}  `json:"mismatches"`
}

func getReplayReport(replayID string) *ReplayReport {
	resp, err := http.Get(baseURL + "/api/v1/sandbox/replay/sessions/" + replayID + "/report")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["code"] != "0000" {
		return nil
	}

	data := result["data"].(map[string]interface{})
	return &ReplayReport{
		TotalRequests:   int(data["totalRequests"].(float64)),
		MatchedCount:    int(data["matchedCount"].(float64)),
		MismatchedCount: int(data["mismatchedCount"].(float64)),
		MatchRate:       data["matchRate"].(float64),
		Mismatches:      data["mismatches"].([]interface{}),
	}
}
