package validation

import (
	"strings"

	"risk-decision-engine/internal/api/errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Validate 验证请求参数
func Validate(c *gin.Context, obj interface{}) *errors.AppError {
	if err := c.ShouldBindJSON(obj); err != nil {
		return errors.New(errors.CodeInvalidParams, "请求参数解析失败: "+err.Error())
	}

	if err := validate.Struct(obj); err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			errMsgs := make([]string, 0, len(validationErrs))
			for _, e := range validationErrs {
				errMsgs = append(errMsgs, formatValidationError(e))
			}
			return errors.New(errors.CodeInvalidParams, strings.Join(errMsgs, "; "))
		}
		return errors.New(errors.CodeInvalidParams, "参数验证失败")
	}

	return nil
}

// formatValidationError 格式化验证错误
func formatValidationError(e validator.FieldError) string {
	field := e.Field()
	switch e.Tag() {
	case "required":
		return field + " 是必填项"
	case "email":
		return field + " 格式不正确"
	case "min":
		return field + " 最小值为 " + e.Param()
	case "max":
		return field + " 最大值为 " + e.Param()
	case "len":
		return field + " 长度必须为 " + e.Param()
	case "gte":
		return field + " 必须大于等于 " + e.Param()
	case "lte":
		return field + " 必须小于等于 " + e.Param()
	case "gt":
		return field + " 必须大于 " + e.Param()
	case "lt":
		return field + " 必须小于 " + e.Param()
	default:
		return field + " 验证失败，条件: " + e.Tag()
	}
}

// RegisterValidation 注册自定义验证规则
func RegisterValidation(tag string, fn validator.Func) error {
	return validate.RegisterValidation(tag, fn)
}

// GetValidator 获取验证器实例
func GetValidator() *validator.Validate {
	return validate
}
