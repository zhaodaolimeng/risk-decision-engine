package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode string

// 错误码定义
const (
	// 通用错误码 (1xxx)
	CodeSuccess           ErrorCode = "0000"
	CodeInternalError     ErrorCode = "1000"
	CodeInvalidParams     ErrorCode = "1001"
	CodeUnauthorized      ErrorCode = "1002"
	CodeForbidden         ErrorCode = "1003"
	CodeNotFound          ErrorCode = "1004"
	CodeTooManyRequests   ErrorCode = "1005"

	// 规则引擎错误码 (2xxx)
	CodeRuleLoadError     ErrorCode = "2001"
	CodeRuleExecuteError  ErrorCode = "2002"
	CodeRuleInvalid       ErrorCode = "2003"

	// 模型服务错误码 (3xxx)
	CodeModelCallError    ErrorCode = "3001"
	CodeModelTimeout      ErrorCode = "3002"
	CodeModelInvalidResp  ErrorCode = "3003"

	// 数据源错误码 (4xxx)
	CodeDataSourceError   ErrorCode = "4001"
	CodeDataSourceTimeout ErrorCode = "4002"

	// 决策流错误码 (5xxx)
	CodeFlowLoadError     ErrorCode = "5001"
	CodeFlowExecuteError  ErrorCode = "5002"
)

// ErrorCodeMessage 错误码消息映射
var ErrorCodeMessage = map[ErrorCode]string{
	CodeSuccess:           "成功",
	CodeInternalError:     "内部服务器错误",
	CodeInvalidParams:     "请求参数错误",
	CodeUnauthorized:      "未授权",
	CodeForbidden:         "禁止访问",
	CodeNotFound:          "资源不存在",
	CodeTooManyRequests:   "请求过于频繁",
	CodeRuleLoadError:     "规则加载失败",
	CodeRuleExecuteError:  "规则执行失败",
	CodeRuleInvalid:       "规则配置无效",
	CodeModelCallError:    "模型服务调用失败",
	CodeModelTimeout:      "模型服务超时",
	CodeModelInvalidResp:  "模型响应格式错误",
	CodeDataSourceError:   "数据源调用失败",
	CodeDataSourceTimeout: "数据源超时",
	CodeFlowLoadError:     "决策流加载失败",
	CodeFlowExecuteError:  "决策流执行失败",
}

// ErrorCodeHTTPStatus 错误码HTTP状态码映射
var ErrorCodeHTTPStatus = map[ErrorCode]int{
	CodeSuccess:           http.StatusOK,
	CodeInternalError:     http.StatusInternalServerError,
	CodeInvalidParams:     http.StatusBadRequest,
	CodeUnauthorized:      http.StatusUnauthorized,
	CodeForbidden:         http.StatusForbidden,
	CodeNotFound:          http.StatusNotFound,
	CodeTooManyRequests:   http.StatusTooManyRequests,
	CodeRuleLoadError:     http.StatusInternalServerError,
	CodeRuleExecuteError:  http.StatusInternalServerError,
	CodeRuleInvalid:       http.StatusBadRequest,
	CodeModelCallError:    http.StatusBadGateway,
	CodeModelTimeout:      http.StatusGatewayTimeout,
	CodeModelInvalidResp:  http.StatusBadGateway,
	CodeDataSourceError:   http.StatusBadGateway,
	CodeDataSourceTimeout: http.StatusGatewayTimeout,
	CodeFlowLoadError:     http.StatusInternalServerError,
	CodeFlowExecuteError:  http.StatusInternalServerError,
}

// AppError 应用错误
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Detail  string    `json:"detail,omitempty"`
	Err     error     `json:"-"`
}

// New 创建新的应用错误
func New(code ErrorCode, detail ...string) *AppError {
	msg := ErrorCodeMessage[code]
	if msg == "" {
		msg = "未知错误"
	}

	err := &AppError{
		Code:    code,
		Message: msg,
	}

	if len(detail) > 0 && detail[0] != "" {
		err.Detail = detail[0]
	}

	return err
}

// Wrap 包装原始错误
func Wrap(code ErrorCode, err error, detail ...string) *AppError {
	appErr := New(code, detail...)
	appErr.Err = err
	if err != nil {
		if appErr.Detail == "" {
			appErr.Detail = err.Error()
		} else {
			appErr.Detail = fmt.Sprintf("%s: %v", appErr.Detail, err)
		}
	}
	return appErr
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	if e.Detail != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现errors.Unwrap接口
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatus 获取HTTP状态码
func (e *AppError) HTTPStatus() int {
	if status, ok := ErrorCodeHTTPStatus[e.Code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// Is 判断错误码
func (e *AppError) Is(code ErrorCode) bool {
	return e.Code == code
}

// Response 统一响应结构
type Response struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Detail  string      `json:"detail,omitempty"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Code:    CodeSuccess,
		Message: ErrorCodeMessage[CodeSuccess],
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(err *AppError) *Response {
	return &Response{
		Code:    err.Code,
		Message: err.Message,
		Detail:  err.Detail,
	}
}

// NewErrorResponseFromCode 从错误码创建错误响应
func NewErrorResponseFromCode(code ErrorCode, detail ...string) *Response {
	return NewErrorResponse(New(code, detail...))
}

// IsAppError 判断是否为AppError
func IsAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	appErr, ok := err.(*AppError)
	return appErr, ok
}

// FromError 从error转换为AppError
func FromError(err error) *AppError {
	if err == nil {
		return nil
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return Wrap(CodeInternalError, err)
}
