package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ModelService 模型服务接口
type ModelService interface {
	Predict(modelID string, req *ModelRequest) (*ModelResult, error)
}

// ModelClient 模型服务客户端
type ModelClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// ModelRequest 模型请求
type ModelRequest struct {
	ContractID  string                 `json:"contract_id"`
	Applicant   map[string]interface{} `json:"applicant"`
	Application map[string]interface{} `json:"application"`
}

// ModelResponse 模型响应
type ModelResponse struct {
	Code    interface{} `json:"code"`
	Message string      `json:"message"`
	Data    struct {
		DefaultProbability float64 `json:"default_probability"`
		RiskScore          int     `json:"risk_score"`
		RiskLevel          string  `json:"risk_level"`
	} `json:"data"`
}

// ModelResult 模型执行结果（用于决策上下文）
type ModelResult struct {
	ModelID            string  `json:"modelId"`
	DefaultProbability float64 `json:"defaultProbability"`
	RiskScore          int     `json:"riskScore"`
	RiskLevel          string  `json:"riskLevel"`
}

// NewModelClient 创建模型客户端
func NewModelClient(baseURL, apiKey string) *ModelClient {
	return &ModelClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

// Predict 调用模型预测
func (c *ModelClient) Predict(modelID string, req *ModelRequest) (*ModelResult, error) {
	start := time.Now()
	defer func() {
		fmt.Printf("[模型调用] %s 耗时: %v\n", modelID, time.Since(start))
	}()

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	fmt.Printf("[模型请求] modelId=%s, contractId=%s\n", modelID, req.ContractID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call model service: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("model service error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var modelResp ModelResponse
	if err := json.Unmarshal(respBody, &modelResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// 检查响应码，支持字符串和数字
	codeOK := false
	switch code := modelResp.Code.(type) {
	case string:
		codeOK = code == "0000" || code == "0"
	case float64:
		codeOK = code == 0
	case int:
		codeOK = code == 0
	}

	if !codeOK {
		return nil, fmt.Errorf("model error: code=%v, message=%s", modelResp.Code, modelResp.Message)
	}

	result := &ModelResult{
		ModelID:            modelID,
		DefaultProbability: modelResp.Data.DefaultProbability,
		RiskScore:          modelResp.Data.RiskScore,
		RiskLevel:          modelResp.Data.RiskLevel,
	}

	fmt.Printf("[模型响应] riskScore=%d, riskLevel=%s\n", result.RiskScore, result.RiskLevel)

	return result, nil
}
