package service

import (
	"time"

	"risk-decision-engine/pkg/cache"
	"risk-decision-engine/pkg/config"
)

const (
	// 缓存Key前缀
	ruleCachePrefix      = "rule:"
	decisionCachePrefix  = "decision:"
	hotDataCachePrefix   = "hot:"
	counterCachePrefix   = "counter:"

	// 默认TTL
	defaultCacheTTL      = 5 * time.Minute
	ruleConfigCacheTTL   = 10 * time.Minute
	hotDataCacheTTL      = 1 * time.Minute
)

// CacheService 缓存服务
type CacheService struct {
	cfg *config.Config
}

// NewCacheService 创建缓存服务
func NewCacheService(cfg *config.Config) *CacheService {
	return &CacheService{cfg: cfg}
}

// RuleConfigCache 规则配置缓存
type RuleConfigCache struct {
	RuleID      string
	Version     string
	ConfigJSON  string
	UpdatedAt   time.Time
}

// SetRuleConfig 缓存规则配置
func (s *CacheService) SetRuleConfig(ruleID string, data *RuleConfigCache) error {
	key := ruleCachePrefix + ruleID
	ttl := ruleConfigCacheTTL
	if s.cfg.Strategy.CacheTTL > 0 {
		ttl = s.cfg.Strategy.CacheTTL
	}
	return cache.Set(key, data, ttl)
}

// GetRuleConfig 获取规则配置缓存
func (s *CacheService) GetRuleConfig(ruleID string) (*RuleConfigCache, error) {
	key := ruleCachePrefix + ruleID
	var data RuleConfigCache
	err := cache.Get(key, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// DeleteRuleConfig 删除规则配置缓存
func (s *CacheService) DeleteRuleConfig(ruleID string) error {
	key := ruleCachePrefix + ruleID
	return cache.Delete(key)
}

// DeleteAllRuleConfigs 删除所有规则配置缓存
func (s *CacheService) DeleteAllRuleConfigs() error {
	return cache.DeleteByPattern(ruleCachePrefix + "*")
}

// DecisionCache 决策结果缓存
type DecisionCache struct {
	RequestID     string
	Decision      string
	DecisionCode  string
	DecisionReason string
	ExecutedAt    time.Time
}

// SetDecision 缓存决策结果
func (s *CacheService) SetDecision(requestID string, data *DecisionCache) error {
	key := decisionCachePrefix + requestID
	return cache.Set(key, data, defaultCacheTTL)
}

// GetDecision 获取决策缓存
func (s *CacheService) GetDecision(requestID string) (*DecisionCache, error) {
	key := decisionCachePrefix + requestID
	var data DecisionCache
	err := cache.Get(key, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// HasDecision 检查决策是否已缓存
func (s *CacheService) HasDecision(requestID string) (bool, error) {
	key := decisionCachePrefix + requestID
	return cache.Exists(key)
}

// HotDataCache 热点数据缓存
func (s *CacheService) SetHotData(key string, data interface{}) error {
	cacheKey := hotDataCachePrefix + key
	return cache.Set(cacheKey, data, hotDataCacheTTL)
}

func (s *CacheService) GetHotData(key string, dest interface{}) error {
	cacheKey := hotDataCachePrefix + key
	return cache.Get(cacheKey, dest)
}

func (s *CacheService) DeleteHotData(key string) error {
	cacheKey := hotDataCachePrefix + key
	return cache.Delete(cacheKey)
}

// IncrementCounter 计数器自增
func (s *CacheService) IncrementCounter(name string) (int64, error) {
	key := counterCachePrefix + name
	return cache.Incr(key)
}

// GetCounter 获取计数器
func (s *CacheService) GetCounter(name string) (int64, error) {
	key := counterCachePrefix + name
	var count int64
	err := cache.Get(key, &count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// SetCounterExpire 设置计数器过期时间
func (s *CacheService) SetCounterExpire(name string, ttl time.Duration) error {
	key := counterCachePrefix + name
	return cache.Expire(key, ttl)
}

// DeleteCounter 删除计数器
func (s *CacheService) DeleteCounter(name string) error {
	key := counterCachePrefix + name
	return cache.Delete(key)
}

// RecordDecisionHit 记录决策缓存命中
func (s *CacheService) RecordDecisionHit() (int64, error) {
	return s.IncrementCounter("decision:hit")
}

// RecordDecisionMiss 记录决策缓存未命中
func (s *CacheService) RecordDecisionMiss() (int64, error) {
	return s.IncrementCounter("decision:miss")
}

// GetDecisionCacheStats 获取决策缓存统计
func (s *CacheService) GetDecisionCacheStats() (hit, miss int64, err error) {
	hit, _ = s.GetCounter("decision:hit")
	miss, _ = s.GetCounter("decision:miss")
	return hit, miss, nil
}

// ClearAll 清空所有缓存
func (s *CacheService) ClearAll() error {
	if err := cache.DeleteByPattern(ruleCachePrefix + "*"); err != nil {
		return err
	}
	if err := cache.DeleteByPattern(decisionCachePrefix + "*"); err != nil {
		return err
	}
	if err := cache.DeleteByPattern(hotDataCachePrefix + "*"); err != nil {
		return err
	}
	if err := cache.DeleteByPattern(counterCachePrefix + "*"); err != nil {
		return err
	}
	return nil
}
