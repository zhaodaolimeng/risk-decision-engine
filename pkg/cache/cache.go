package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"risk-decision-engine/pkg/config"
	"risk-decision-engine/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var (
	globalRedis *redis.Client
	ctx         = context.Background()
)

// Init 初始化Redis
func Init(cfg *config.RedisConfig) error {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}

	globalRedis = client
	logger.Info("redis connected")

	return nil
}

// Client 获取Redis客户端
func Client() *redis.Client {
	if globalRedis == nil {
		panic("redis not initialized")
	}
	return globalRedis
}

// Close 关闭Redis连接
func Close() error {
	if globalRedis != nil {
		return globalRedis.Close()
	}
	return nil
}

// Set 设置缓存
func Set(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return Client().Set(ctx, key, data, ttl).Err()
}

// Get 获取缓存
func Get(key string, dest interface{}) error {
	data, err := Client().Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found")
		}
		return err
	}
	return json.Unmarshal(data, dest)
}

// Delete 删除缓存
func Delete(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return Client().Del(ctx, keys...).Err()
}

// DeleteByPattern 按模式删除缓存
func DeleteByPattern(pattern string) error {
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = Client().Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := Client().Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		if cursor == 0 {
			break
		}
	}
	return nil
}

// Exists 检查key是否存在
func Exists(key string) (bool, error) {
	count, err := Client().Exists(ctx, key).Result()
	return count > 0, err
}

// Expire 设置过期时间
func Expire(key string, ttl time.Duration) error {
	return Client().Expire(ctx, key, ttl).Err()
}

// TTL 获取剩余时间
func TTL(key string) (time.Duration, error) {
	return Client().TTL(ctx, key).Result()
}

// Incr 自增
func Incr(key string) (int64, error) {
	return Client().Incr(ctx, key).Result()
}

// Decr 自减
func Decr(key string) (int64, error) {
	return Client().Decr(ctx, key).Result()
}

// HSet 设置Hash
func HSet(key string, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return Client().HSet(ctx, key, field, data).Err()
}

// HGet 获取Hash
func HGet(key string, field string, dest interface{}) error {
	data, err := Client().HGet(ctx, key, field).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("field not found")
		}
		return err
	}
	return json.Unmarshal(data, dest)
}

// HGetAll 获取所有Hash字段
func HGetAll(key string) (map[string]string, error) {
	return Client().HGetAll(ctx, key).Result()
}

// HDelete 删除Hash字段
func HDelete(key string, fields ...string) error {
	return Client().HDel(ctx, key, fields...).Err()
}

// LPush 左推入列表
func LPush(key string, values ...interface{}) error {
	return Client().LPush(ctx, key, values...).Err()
}

// RPush 右推入列表
func RPush(key string, values ...interface{}) error {
	return Client().RPush(ctx, key, values...).Err()
}

// LPop 左弹出列表
func LPop(key string) (string, error) {
	return Client().LPop(ctx, key).Result()
}

// RPop 右弹出列表
func RPop(key string) (string, error) {
	return Client().RPop(ctx, key).Result()
}

// LLen 获取列表长度
func LLen(key string) (int64, error) {
	return Client().LLen(ctx, key).Result()
}

// LRange 获取列表范围
func LRange(key string, start, stop int64) ([]string, error) {
	return Client().LRange(ctx, key, start, stop).Result()
}
