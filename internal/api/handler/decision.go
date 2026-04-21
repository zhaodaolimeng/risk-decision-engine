package handler

import (
	"net/http"
	"time"

	"risk-decision-engine/internal/api/dto"
	"risk-decision-engine/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ExecuteDecision 执行决策
func ExecuteDecision(c *gin.Context) {
	var req dto.DecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("invalid request: %v", err)
		c.JSON(http.StatusBadRequest, dto.Error("1001", "参数错误"))
		return
	}

	// 生成决策ID
	decisionID := generateDecisionID()
	startTime := time.Now()

	logger.Infof("execute decision, requestId=%s, businessId=%s, decisionId=%s",
		req.RequestID, req.BusinessID, decisionID)

	// TODO: 实际决策逻辑
	// 1. 获取策略
	// 2. 预加载数据
	// 3. 执行决策流
	// 4. 返回结果

	// 模拟响应
	resp := &dto.DecisionResponse{
		DecisionID:     decisionID,
		BusinessID:     req.BusinessID,
		Decision:       "APPROVE",
		DecisionCode:   "APPROVE_001",
		DecisionReason: "综合评估通过",
		Score:          720,
		RiskLevel:      "LOW",
		ExecuteTime:    time.Since(startTime).Milliseconds(),
	}

	c.JSON(http.StatusOK, dto.SuccessWithRequestID(req.RequestID, resp))
}

// QueryDecision 查询决策
func QueryDecision(c *gin.Context) {
	var req dto.DecisionQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error("1001", "参数错误"))
		return
	}

	if req.DecisionID == "" && req.BusinessID == "" {
		c.JSON(http.StatusBadRequest, dto.Error("1001", "decisionId 或 businessId 必须提供一个"))
		return
	}

	// TODO: 查询决策逻辑
	c.JSON(http.StatusOK, dto.Success(nil))
}

func generateDecisionID() string {
	return "dec_" + uuid.NewString()[:8]
}
