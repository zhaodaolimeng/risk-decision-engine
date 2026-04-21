package dto

import (
	"time"
)

// Response 通用响应
type Response struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	RequestID string      `json:"requestId,omitempty"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(data interface{}) *Response {
	return &Response{
		Code:      "0000",
		Message:   "成功",
		Timestamp: time.Now().UnixMilli(),
		Data:      data,
	}
}

// SuccessWithRequestID 带请求ID的成功响应
func SuccessWithRequestID(requestID string, data interface{}) *Response {
	return &Response{
		Code:      "0000",
		Message:   "成功",
		RequestID: requestID,
		Timestamp: time.Now().UnixMilli(),
		Data:      data,
	}
}

// Error 错误响应
func Error(code, message string) *Response {
	return &Response{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UnixMilli(),
	}
}

// ErrorWithRequestID 带请求ID的错误响应
func ErrorWithRequestID(requestID, code, message string) *Response {
	return &Response{
		Code:      code,
		Message:   message,
		RequestID: requestID,
		Timestamp: time.Now().UnixMilli(),
	}
}

// DecisionRequest 决策请求
type DecisionRequest struct {
	RequestID   string                 `json:"requestId" binding:"required"`
	ProductID   string                 `json:"productId" binding:"required"`
	StrategyID  string                 `json:"strategyId,omitempty"`
	BusinessID  string                 `json:"businessId" binding:"required"`
	ContractID  string                 `json:"contractId,omitempty"`
	CallbackURL string                 `json:"callbackUrl,omitempty"`
	Data        map[string]interface{} `json:"data" binding:"required"`
	Extensions  map[string]interface{} `json:"extensions,omitempty"`
}

// DecisionResponse 决策响应
type DecisionResponse struct {
	DecisionID      string                 `json:"decisionId"`
	BusinessID      string                 `json:"businessId"`
	Decision        string                 `json:"decision"`
	DecisionCode    string                 `json:"decisionCode,omitempty"`
	DecisionReason  string                 `json:"decisionReason"`
	Score           int                    `json:"score,omitempty"`
	RiskLevel       string                 `json:"riskLevel,omitempty"`
	ApprovalAmount  float64                `json:"approvalAmount,omitempty"`
	ApprovalTerm    int                    `json:"approvalTerm,omitempty"`
	InterestRate    float64                `json:"interestRate,omitempty"`
	ModelResults    map[string]interface{} `json:"modelResults,omitempty"`
	ExecuteDetail   interface{}            `json:"executeDetail,omitempty"`
	Suggestions     []string               `json:"suggestions,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	ExecuteTime     int64                  `json:"executeTime"`
}

// DecisionQueryRequest 决策查询请求
type DecisionQueryRequest struct {
	DecisionID string `form:"decisionId"`
	BusinessID string `form:"businessId"`
}
