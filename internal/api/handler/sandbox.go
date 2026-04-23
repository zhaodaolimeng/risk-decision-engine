package handler

import (
	"net/http"
	"time"

	"risk-decision-engine/internal/api/errors"
	"risk-decision-engine/internal/api/middleware"
	"risk-decision-engine/internal/engine/rule"
	"risk-decision-engine/internal/sandbox"

	"github.com/gin-gonic/gin"
)

// SandboxHandler 沙箱API处理器
type SandboxHandler struct {
	recorder   *sandbox.Recorder
	replayer   *sandbox.Replayer
	comparator *sandbox.RuleComparator
	executor   sandbox.DecisionExecutor
}

// NewSandboxHandler 创建沙箱处理器
func NewSandboxHandler(
	recorder *sandbox.Recorder,
	replayer *sandbox.Replayer,
	comparator *sandbox.RuleComparator,
	executor sandbox.DecisionExecutor,
) *SandboxHandler {
	return &SandboxHandler{
		recorder:   recorder,
		replayer:   replayer,
		comparator: comparator,
		executor:   executor,
	}
}

// RegisterRoutes 注册路由
func (h *SandboxHandler) RegisterRoutes(r *gin.Engine) {
	sandbox := r.Group("/api/v1/sandbox")
	{
		// 记录相关
		sandbox.POST("/record/start", h.StartRecording)
		sandbox.POST("/record/stop", h.StopRecording)
		sandbox.GET("/record/sessions", h.ListRecordingSessions)
		sandbox.GET("/record/sessions/:id", h.GetRecordingSession)
		sandbox.GET("/record/sessions/:id/records", h.GetRecordingRecords)

		// 回放相关
		sandbox.POST("/replay/start", h.StartReplay)
		sandbox.POST("/replay/start/options", h.StartReplayWithOptions) // 带选项的回放
		sandbox.GET("/replay/sessions", h.ListReplaySessions)
		sandbox.GET("/replay/sessions/:id", h.GetReplaySession)
		sandbox.GET("/replay/sessions/:id/report", h.GetReplayReport)

		// 规则比对相关
		sandbox.POST("/diff/rules", h.CompareRuleConfigs)
	}
}

// StartRecording godoc
// @Summary 开始流量记录
// @Description 开始新的流量记录会话
// @Tags 沙盒-记录
// @Accept json
// @Produce json
// @Param request body object{name=string} true "会话名称"
// @Success 200 {object} object{code=string,message=string,data=sandbox.RecordingSession}
// @Failure 400 {object} object{code=string,message=string}
// @Failure 500 {object} object{code=string,message=string}
// @Router /api/v1/sandbox/record/start [post]
func (h *SandboxHandler) StartRecording(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithInvalidParams(c, err.Error())
		return
	}

	session, err := h.recorder.StartSession(req.Name)
	if err != nil {
		middleware.AbortWithError(c, errors.Wrap(errors.CodeInternalError, err))
		return
	}

	middleware.RespondSuccess(c, session)
}

// StopRecording godoc
// @Summary 停止流量记录
// @Description 停止当前的流量记录会话
// @Tags 沙盒-记录
// @Accept json
// @Produce json
// @Success 200 {object} object{code=string,message=string,data=sandbox.RecordingSession}
// @Failure 500 {object} object{code=string,message=string}
// @Router /api/v1/sandbox/record/stop [post]
func (h *SandboxHandler) StopRecording(c *gin.Context) {
	session, err := h.recorder.StopSession()
	if err != nil {
		middleware.AbortWithError(c, errors.Wrap(errors.CodeInternalError, err))
		return
	}

	middleware.RespondSuccess(c, session)
}

// ListRecordingSessions godoc
// @Summary 列出记录会话
// @Description 获取所有的记录会话列表
// @Tags 沙盒-记录
// @Accept json
// @Produce json
// @Success 200 {object} object{code=string,message=string,data=object{sessions=[]sandbox.RecordingSession,total=int}}
// @Router /api/v1/sandbox/record/sessions [get]
func (h *SandboxHandler) ListRecordingSessions(c *gin.Context) {
	sessions := h.recorder.ListSessions()
	middleware.RespondSuccess(c, gin.H{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// GetRecordingSession godoc
// @Summary 获取记录会话详情
// @Description 获取指定会话ID的详情
// @Tags 沙盒-记录
// @Accept json
// @Produce json
// @Param id path string true "会话ID"
// @Success 200 {object} object{code=string,message=string,data=sandbox.RecordingSession}
// @Failure 404 {object} object{code=string,message=string}
// @Router /api/v1/sandbox/record/sessions/{id} [get]
func (h *SandboxHandler) GetRecordingSession(c *gin.Context) {
	sessionID := c.Param("id")
	session, err := h.recorder.GetSession(sessionID)
	if err != nil {
		middleware.AbortWithError(c, errors.Wrap(errors.CodeNotFound, err))
		return
	}

	middleware.RespondSuccess(c, session)
}

// GetRecordingRecords godoc
// @Summary 获取记录数据
// @Description 获取指定会话的所有记录数据
// @Tags 沙盒-记录
// @Accept json
// @Produce json
// @Param id path string true "会话ID"
// @Success 200 {object} object{code=string,message=string,data=object{records=[]sandbox.TrafficRecord,total=int}}
// @Failure 404 {object} object{code=string,message=string}
// @Router /api/v1/sandbox/record/sessions/{id}/records [get]
func (h *SandboxHandler) GetRecordingRecords(c *gin.Context) {
	sessionID := c.Param("id")
	records, err := h.recorder.GetRecords(sessionID)
	if err != nil {
		middleware.AbortWithError(c, errors.Wrap(errors.CodeNotFound, err))
		return
	}

	middleware.RespondSuccess(c, gin.H{
		"records": records,
		"total":   len(records),
	})
}

// StartReplay godoc
// @Summary 开始流量回放
// @Description 使用记录会话开始流量回放
// @Tags 沙盒-回放
// @Accept json
// @Produce json
// @Param request body object{name=string,recordingId=string,configPath=string} true "回放请求"
// @Success 200 {object} object{code=string,message=string,data=sandbox.ReplaySession}
// @Failure 400 {object} object{code=string,message=string}
// @Failure 500 {object} object{code=string,message=string}
// @Router /api/v1/sandbox/replay/start [post]
func (h *SandboxHandler) StartReplay(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		RecordingID string `json:"recordingId" binding:"required"`
		ConfigPath  string `json:"configPath"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithInvalidParams(c, err.Error())
		return
	}

	session, err := h.replayer.StartReplay(req.Name, req.RecordingID, req.ConfigPath, h.executor)
	if err != nil {
		middleware.AbortWithError(c, errors.Wrap(errors.CodeInternalError, err))
		return
	}

	middleware.RespondSuccess(c, session)
}

// StartReplayWithOptions godoc
// @Summary 开始流量回放（带选项）
// @Description 使用指定配置和时间范围开始流量回放
// @Tags 沙盒-回放
// @Accept json
// @Produce json
// @Param request body object{name=string,recordingId=string,ruleConfigPath=string,timeFrom=string,timeTo=string,maxParallel=int,dryRun=bool} true "回放请求"
// @Success 200 {object} object{code=string,message=string,data=sandbox.ReplaySession}
// @Failure 400 {object} object{code=string,message=string}
// @Failure 500 {object} object{code=string,message=string}
// @Router /api/v1/sandbox/replay/start/options [post]
func (h *SandboxHandler) StartReplayWithOptions(c *gin.Context) {
	var req struct {
		Name           string  `json:"name" binding:"required"`
		RecordingID    string  `json:"recordingId" binding:"required"`
		RuleConfigPath string  `json:"ruleConfigPath"`
		TimeFrom       *string `json:"timeFrom"` // RFC3339格式
		TimeTo         *string `json:"timeTo"`   // RFC3339格式
		MaxParallel    int     `json:"maxParallel"`
		DryRun         bool    `json:"dryRun"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithInvalidParams(c, err.Error())
		return
	}

	// 解析时间
	var timeFrom, timeTo *time.Time
	if req.TimeFrom != nil && *req.TimeFrom != "" {
		t, err := time.Parse(time.RFC3339, *req.TimeFrom)
		if err != nil {
			middleware.AbortWithInvalidParams(c, "timeFrom格式错误，应为RFC3339格式")
			return
		}
		timeFrom = &t
	}
	if req.TimeTo != nil && *req.TimeTo != "" {
		t, err := time.Parse(time.RFC3339, *req.TimeTo)
		if err != nil {
			middleware.AbortWithInvalidParams(c, "timeTo格式错误，应为RFC3339格式")
			return
		}
		timeTo = &t
	}

	options := &sandbox.ReplayOptions{
		RuleConfigPath: req.RuleConfigPath,
		TimeFrom:       timeFrom,
		TimeTo:         timeTo,
		MaxParallel:    req.MaxParallel,
		DryRun:         req.DryRun,
	}

	session, err := h.replayer.StartReplayWithOptions(req.Name, req.RecordingID, options, h.executor)
	if err != nil {
		middleware.AbortWithError(c, errors.Wrap(errors.CodeInternalError, err))
		return
	}

	middleware.RespondSuccess(c, session)
}

// ListReplaySessions godoc
// @Summary 列出回放会话
// @Description 获取所有的回放会话列表
// @Tags 沙盒-回放
// @Accept json
// @Produce json
// @Success 200 {object} object{code=string,message=string,data=object{sessions=[]sandbox.ReplaySession,total=int}}
// @Router /api/v1/sandbox/replay/sessions [get]
func (h *SandboxHandler) ListReplaySessions(c *gin.Context) {
	sessions := h.replayer.ListSessions()
	middleware.RespondSuccess(c, gin.H{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// GetReplaySession godoc
// @Summary 获取回放会话详情
// @Description 获取指定回放会话ID的详情
// @Tags 沙盒-回放
// @Accept json
// @Produce json
// @Param id path string true "会话ID"
// @Success 200 {object} object{code=string,message=string,data=sandbox.ReplaySession}
// @Failure 404 {object} object{code=string,message=string}
// @Router /api/v1/sandbox/replay/sessions/{id} [get]
func (h *SandboxHandler) GetReplaySession(c *gin.Context) {
	sessionID := c.Param("id")
	session, err := h.replayer.GetSession(sessionID)
	if err != nil {
		middleware.AbortWithError(c, errors.Wrap(errors.CodeNotFound, err))
		return
	}

	middleware.RespondSuccess(c, session)
}

// GetReplayReport godoc
// @Summary 获取回放报告
// @Description 获取指定回放会话的比对报告
// @Tags 沙盒-回放
// @Accept json
// @Produce json
// @Param id path string true "会话ID"
// @Success 200 {object} object{code=string,message=string,data=sandbox.ReplayReport}
// @Failure 404 {object} object{code=string,message=string}
// @Router /api/v1/sandbox/replay/sessions/{id}/report [get]
func (h *SandboxHandler) GetReplayReport(c *gin.Context) {
	sessionID := c.Param("id")
	report, err := h.replayer.GetReport(sessionID)
	if err != nil {
		middleware.AbortWithError(c, errors.Wrap(errors.CodeNotFound, err))
		return
	}

	middleware.RespondSuccess(c, report)
}

// CompareRuleConfigs godoc
// @Summary 比对规则配置
// @Description 比对两个规则配置文件的差异
// @Tags 沙盒-比对
// @Accept json
// @Produce json
// @Param request body object{name=string,oldConfigPath=string,newConfigPath=string} true "比对请求"
// @Success 200 {object} object{code=string,message=string,data=sandbox.ConfigDiffReport}
// @Failure 400 {object} object{code=string,message=string}
// @Failure 500 {object} object{code=string,message=string}
// @Router /api/v1/sandbox/diff/rules [post]
func (h *SandboxHandler) CompareRuleConfigs(c *gin.Context) {
	var req struct {
		Name          string `json:"name" binding:"required"`
		OldConfigPath string `json:"oldConfigPath" binding:"required"`
		NewConfigPath string `json:"newConfigPath" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithInvalidParams(c, err.Error())
		return
	}

	// 加载旧配置
	oldConfig, err := rule.LoadRuleConfigFromFile(req.OldConfigPath)
	if err != nil {
		middleware.AbortWithError(c, errors.Wrap(errors.CodeRuleLoadError, err, "加载旧配置失败"))
		return
	}

	// 加载新配置
	newConfig, err := rule.LoadRuleConfigFromFile(req.NewConfigPath)
	if err != nil {
		middleware.AbortWithError(c, errors.Wrap(errors.CodeRuleLoadError, err, "加载新配置失败"))
		return
	}

	// 执行比对
	report, err := h.comparator.CompareRuleConfigs(req.Name, oldConfig, newConfig)
	if err != nil {
		middleware.AbortWithError(c, errors.Wrap(errors.CodeInternalError, err, "比对失败"))
		return
	}

	middleware.RespondSuccess(c, report)
}
