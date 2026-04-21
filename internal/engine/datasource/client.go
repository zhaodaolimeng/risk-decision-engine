package datasource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DataSourceService 数据源服务接口
type DataSourceService interface {
	Fetch(dsID string, req *DataSourceRequest) (*DataSourceResult, error)
}

// DataSourceClient 数据源客户端
type DataSourceClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// DataSourceRequest 数据源请求
type DataSourceRequest struct {
	ContractID string                 `json:"contract_id"`
	UserID     string                 `json:"user_id"`
	Params     map[string]interface{} `json:"params"`
}

// DataSourceResponse 数据源响应
type DataSourceResponse struct {
	Code    interface{}              `json:"code"`
	Message string                   `json:"message"`
	Data    map[string]interface{}   `json:"data"`
}

// DataSourceResult 数据源执行结果（用于决策上下文）
type DataSourceResult struct {
	DataSourceID string                 `json:"dataSourceId"`
	Data         map[string]interface{} `json:"data"`
}

// NewDataSourceClient 创建数据源客户端
func NewDataSourceClient(baseURL, apiKey string) *DataSourceClient {
	return &DataSourceClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

// Fetch 调用数据源获取数据
func (c *DataSourceClient) Fetch(dsID string, req *DataSourceRequest) (*DataSourceResult, error) {
	start := time.Now()
	defer func() {
		fmt.Printf("[数据源调用] %s 耗时: %v\n", dsID, time.Since(start))
	}()

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s", c.baseURL, dsID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	fmt.Printf("[数据源请求] dsId=%s, userId=%s, contractId=%s\n", dsID, req.UserID, req.ContractID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call datasource: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("datasource error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var dsResp DataSourceResponse
	if err := json.Unmarshal(respBody, &dsResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// 检查响应码，支持字符串和数字
	codeOK := false
	switch code := dsResp.Code.(type) {
	case string:
		codeOK = code == "0000" || code == "0"
	case float64:
		codeOK = code == 0
	case int:
		codeOK = code == 0
	}

	if !codeOK {
		return nil, fmt.Errorf("datasource error: code=%v, message=%s", dsResp.Code, dsResp.Message)
	}

	result := &DataSourceResult{
		DataSourceID: dsID,
		Data:         dsResp.Data,
	}

	fmt.Printf("[数据源响应] dsId=%s, data keys=%v\n", dsID, getMapKeys(result.Data))

	return result, nil
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
