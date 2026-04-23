package metrics

import (
	"sync"
	"time"
)

// 指标名称常量
const (
	MetricDecisionTotal      = "decision.total"
	MetricDecisionApprove    = "decision.approve"
	MetricDecisionReject     = "decision.reject"
	MetricDecisionError      = "decision.error"
	MetricDecisionDuration   = "decision.duration"
	MetricDecisionQPS        = "decision.qps"
	MetricCacheHit           = "cache.hit"
	MetricCacheMiss          = "cache.miss"
	MetricRuleExecTotal      = "rule.exec.total"
	MetricRuleExecTime       = "rule.exec.time"
	MetricAPICalls           = "api.calls"
	MetricAPIErrors          = "api.errors"
	MetricAPILatency         = "api.latency"
)

// Metrics 指标收集器
type Metrics struct {
	mu sync.RWMutex

	// 决策指标
	decisionTotal      int64
	decisionApprove    int64
	decisionReject     int64
	decisionError      int64
	decisionDurations  []time.Duration

	// 缓存指标
	cacheHit           int64
	cacheMiss          int64

	// 规则指标
	ruleExecTotal      int64
	ruleExecDurations  []time.Duration

	// API指标
	apiCalls           map[string]int64
	apiErrors          map[string]int64
	apiLatency         map[string][]time.Duration

	// 时间窗口数据
	qpsData            []int64
	lastResetTime      time.Time
}

var (
	instance *Metrics
	once     sync.Once
)

// Get 获取指标实例
func Get() *Metrics {
	once.Do(func() {
		instance = &Metrics{
			apiCalls:       make(map[string]int64),
			apiErrors:      make(map[string]int64),
			apiLatency:     make(map[string][]time.Duration),
			lastResetTime:  time.Now(),
		}
	})
	return instance
}

// RecordDecision 记录决策
func (m *Metrics) RecordDecision(decision string, duration time.Duration, hasError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.decisionTotal++
	m.decisionDurations = append(m.decisionDurations, duration)
	if len(m.decisionDurations) > 1000 {
		m.decisionDurations = m.decisionDurations[1:]
	}

	switch decision {
	case "APPROVE":
		m.decisionApprove++
	case "REJECT":
		m.decisionReject++
	}

	if hasError {
		m.decisionError++
	}
}

// RecordCacheHit 记录缓存命中
func (m *Metrics) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheHit++
}

// RecordCacheMiss 记录缓存未命中
func (m *Metrics) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheMiss++
}

// RecordRuleExec 记录规则执行
func (m *Metrics) RecordRuleExec(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ruleExecTotal++
	m.ruleExecDurations = append(m.ruleExecDurations, duration)
	if len(m.ruleExecDurations) > 1000 {
		m.ruleExecDurations = m.ruleExecDurations[1:]
	}
}

// RecordAPI 记录API调用
func (m *Metrics) RecordAPI(path string, duration time.Duration, hasError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.apiCalls[path]++
	if hasError {
		m.apiErrors[path]++
	}

	m.apiLatency[path] = append(m.apiLatency[path], duration)
	if len(m.apiLatency[path]) > 100 {
		m.apiLatency[path] = m.apiLatency[path][1:]
	}
}

// GetStats 获取统计数据
func (m *Metrics) GetStats() *Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &Stats{
		DecisionTotal:      m.decisionTotal,
		DecisionApprove:    m.decisionApprove,
		DecisionReject:     m.decisionReject,
		DecisionError:      m.decisionError,
		CacheHit:           m.cacheHit,
		CacheMiss:          m.cacheMiss,
		RuleExecTotal:      m.ruleExecTotal,
		APICalls:           make(map[string]int64),
		APIErrors:          make(map[string]int64),
		Timestamp:          time.Now(),
	}

	// 计算决策相关指标
	if m.decisionTotal > 0 {
		stats.ApproveRate = float64(m.decisionApprove) / float64(m.decisionTotal) * 100
		stats.RejectRate = float64(m.decisionReject) / float64(m.decisionTotal) * 100
		stats.ErrorRate = float64(m.decisionError) / float64(m.decisionTotal) * 100
	}

	// 计算缓存命中率
	totalCache := m.cacheHit + m.cacheMiss
	if totalCache > 0 {
		stats.CacheHitRate = float64(m.cacheHit) / float64(totalCache) * 100
	}

	// 计算决策平均耗时
	if len(m.decisionDurations) > 0 {
		var total time.Duration
		for _, d := range m.decisionDurations {
			total += d
		}
		stats.AvgDecisionDuration = total / time.Duration(len(m.decisionDurations))
	}

	// 计算规则平均耗时
	if len(m.ruleExecDurations) > 0 {
		var total time.Duration
		for _, d := range m.ruleExecDurations {
			total += d
		}
		stats.AvgRuleDuration = total / time.Duration(len(m.ruleExecDurations))
	}

	// API指标
	for k, v := range m.apiCalls {
		stats.APICalls[k] = v
	}
	for k, v := range m.apiErrors {
		stats.APIErrors[k] = v
	}

	// API平均延迟
	stats.APILatency = make(map[string]time.Duration)
	for path, latencies := range m.apiLatency {
		if len(latencies) > 0 {
			var total time.Duration
			for _, l := range latencies {
				total += l
			}
			stats.APILatency[path] = total / time.Duration(len(latencies))
		}
	}

	return stats
}

// Reset 重置指标
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.decisionTotal = 0
	m.decisionApprove = 0
	m.decisionReject = 0
	m.decisionError = 0
	m.decisionDurations = nil
	m.cacheHit = 0
	m.cacheMiss = 0
	m.ruleExecTotal = 0
	m.ruleExecDurations = nil
	m.apiCalls = make(map[string]int64)
	m.apiErrors = make(map[string]int64)
	m.apiLatency = make(map[string][]time.Duration)
	m.lastResetTime = time.Now()
}

// Stats 统计数据
type Stats struct {
	DecisionTotal       int64            `json:"decisionTotal"`
	DecisionApprove     int64            `json:"decisionApprove"`
	DecisionReject      int64            `json:"decisionReject"`
	DecisionError       int64            `json:"decisionError"`
	ApproveRate         float64          `json:"approveRate"`
	RejectRate          float64          `json:"rejectRate"`
	ErrorRate           float64          `json:"errorRate"`
	AvgDecisionDuration time.Duration    `json:"avgDecisionDuration"`
	CacheHit            int64            `json:"cacheHit"`
	CacheMiss           int64            `json:"cacheMiss"`
	CacheHitRate        float64          `json:"cacheHitRate"`
	RuleExecTotal       int64            `json:"ruleExecTotal"`
	AvgRuleDuration     time.Duration    `json:"avgRuleDuration"`
	APICalls            map[string]int64 `json:"apiCalls"`
	APIErrors           map[string]int64 `json:"apiErrors"`
	APILatency          map[string]time.Duration `json:"apiLatency"`
	Timestamp           time.Time        `json:"timestamp"`
}
