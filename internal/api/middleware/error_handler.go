package middleware

import (
	"net/http"
	"runtime/debug"

	"risk-decision-engine/internal/api/errors"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 处理gin.Context中的错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			handleError(c, err.Err)
			return
		}
	}
}

// RecoveryHandler Panic恢复中间件
func RecoveryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				debug.PrintStack()
				appErr := errors.New(errors.CodeInternalError, "系统异常")
				respondWithError(c, appErr)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// handleError 处理错误
func handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	appErr := errors.FromError(err)
	respondWithError(c, appErr)
}

// respondWithError 响应错误
func respondWithError(c *gin.Context, err *errors.AppError) {
	httpStatus := err.HTTPStatus()
	if c.Writer.Written() {
		return
	}
	c.JSON(httpStatus, errors.NewErrorResponse(err))
}

// RespondSuccess 响应成功
func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, errors.NewSuccessResponse(data))
}

// RespondError 响应错误
func RespondError(c *gin.Context, err *errors.AppError) {
	respondWithError(c, err)
}

// AbortWithError 中止并响应错误
func AbortWithError(c *gin.Context, err *errors.AppError) {
	respondWithError(c, err)
	c.Abort()
}

// AbortWithInvalidParams 中止并响应参数错误
func AbortWithInvalidParams(c *gin.Context, detail string) {
	AbortWithError(c, errors.New(errors.CodeInvalidParams, detail))
}

// AbortWithInternalError 中止并响应内部错误
func AbortWithInternalError(c *gin.Context, err error) {
	AbortWithError(c, errors.Wrap(errors.CodeInternalError, err))
}
