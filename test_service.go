package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TestCase struct {
	name     string
	age      int
	expected string
}

func main() {
	fmt.Println("========================================")
	fmt.Println("   风险决策引擎 - 自测")
	fmt.Println("========================================")
	fmt.Println()

	testCases := []TestCase{
		{"年龄 20 岁", 20, "REJECT"},
		{"年龄 25 岁", 25, "APPROVE"},
		{"年龄 60 岁", 60, "APPROVE"},
		{"年龄 61 岁", 61, "REJECT"},
	}

	baseURL := "http://127.0.0.1:8080"

	fmt.Println("等待服务启动...")
	time.Sleep(2 * time.Second)

	// 健康检查
	fmt.Println("\n1. 健康检查")
	healthResp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("   ✗ 失败: %v\n", err)
		return
	}
	defer healthResp.Body.Close()

	if healthResp.StatusCode == 200 {
		fmt.Println("   ✓ 通过")
	} else {
		fmt.Printf("   ✗ 状态码: %d\n", healthResp.StatusCode)
	}

	// 测试决策接口
	fmt.Println("\n2. 决策接口测试")
	fmt.Println("----------------")

	allPassed := true
	for _, tc := range testCases {
		fmt.Printf("\n测试: %s\n", tc.name)

		reqBody := map[string]interface{}{
			"requestId":  "test_" + tc.name,
			"businessId": "biz_001",
			"data": map[string]interface{}{
				"applicant": map[string]interface{}{
					"age": tc.age,
				},
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		resp, err := http.Post(
			baseURL+"/api/v1/decision/execute",
			"application/json",
			bytes.NewReader(jsonBody),
		)

		if err != nil {
			fmt.Printf("   ✗ 请求失败: %v\n", err)
			allPassed = false
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var result map[string]interface{}
		json.Unmarshal(body, &result)

		var actualDecision string
		if data, ok := result["data"].(map[string]interface{}); ok {
			if d, ok := data["decision"]; ok {
				actualDecision = fmt.Sprintf("%v", d)
			}
			if reason, ok := data["decisionReason"]; ok {
				fmt.Printf("   原因: %v\n", reason)
			}
		}

		if actualDecision == tc.expected {
			fmt.Printf("   ✓ 决策: %s (期望: %s)\n", actualDecision, tc.expected)
		} else {
			fmt.Printf("   ✗ 决策: %s (期望: %s)\n", actualDecision, tc.expected)
			allPassed = false
		}
	}

	fmt.Println("\n----------------")
	if allPassed {
		fmt.Println("✓ 所有测试通过!")
	} else {
		fmt.Println("✗ 部分测试失败")
	}
	fmt.Println("================================")
}
