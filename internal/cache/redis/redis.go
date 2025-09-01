package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"taskmanage/internal/cache"
	"taskmanage/internal/config"
	"taskmanage/pkg/logger"
)

// RedisCache Redis缓存实现
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(cfg *config.Config) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.GetRedisAddr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.Database,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConn,
		MaxRetries:   cfg.Redis.MaxRetries,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
		IdleTimeout:  5 * time.Minute,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis连接失败: %w", err)
	}

	return &RedisCache{
		client: rdb,
		prefix: "taskmanage:",
	}, nil
}

// buildKey 构建带前缀的键
func (r *RedisCache) buildKey(key string) string {
	return r.prefix + key
}

// Get 获取缓存
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	result, err := r.client.Get(ctx, r.buildKey(key)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, cache.ErrCacheKeyNotFound
		}
		logger.Errorf("Redis GET失败: %v", err)
		return nil, cache.ErrCacheConnection.WithCause(err)
	}

	return result, nil
}

// Set 设置缓存
func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := r.client.Set(ctx, r.buildKey(key), value, ttl).Err(); err != nil {
		logger.Errorf("Redis SET失败: %v", err)
		return cache.ErrCacheConnection.WithCause(err)
	}

	return nil
}

// Delete 删除缓存
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := r.client.Del(ctx, r.buildKey(key)).Err(); err != nil {
		logger.Errorf("Redis DEL失败: %v", err)
		return cache.ErrCacheConnection.WithCause(err)
	}

	return nil
}

// Exists 检查键是否存在
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	count, err := r.client.Exists(ctx, r.buildKey(key)).Result()
	if err != nil {
		logger.Errorf("Redis EXISTS失败: %v", err)
		return false, cache.ErrCacheConnection.WithCause(err)
	}

	return count > 0, nil
}

// MGet 批量获取
func (r *RedisCache) MGet(ctx context.Context, keys []string) (map[string][]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(keys) == 0 {
		return make(map[string][]byte), nil
	}

	// 构建带前缀的键
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = r.buildKey(key)
	}

	values, err := r.client.MGet(ctx, prefixedKeys...).Result()
	if err != nil {
		logger.Errorf("Redis MGET失败: %v", err)
		return nil, cache.ErrCacheConnection.WithCause(err)
	}

	result := make(map[string][]byte)
	for i, value := range values {
		if value != nil {
			if str, ok := value.(string); ok {
				result[keys[i]] = []byte(str)
			}
		}
	}

	return result, nil
}

// MSet 批量设置
func (r *RedisCache) MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if len(items) == 0 {
		return nil
	}

	// 使用Pipeline提高性能
	pipe := r.client.Pipeline()

	for key, value := range items {
		pipe.Set(ctx, r.buildKey(key), value, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Errorf("Redis MSET失败: %v", err)
		return cache.ErrCacheConnection.WithCause(err)
	}

	return nil
}

// MDelete 批量删除
func (r *RedisCache) MDelete(ctx context.Context, keys []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(keys) == 0 {
		return nil
	}

	// 构建带前缀的键
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = r.buildKey(key)
	}

	if err := r.client.Del(ctx, prefixedKeys...).Err(); err != nil {
		logger.Errorf("Redis MDEL失败: %v", err)
		return cache.ErrCacheConnection.WithCause(err)
	}

	return nil
}

// Increment 递增
func (r *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	result, err := r.client.IncrBy(ctx, r.buildKey(key), delta).Result()
	if err != nil {
		logger.Errorf("Redis INCRBY失败: %v", err)
		return 0, cache.ErrCacheConnection.WithCause(err)
	}

	return result, nil
}

// Decrement 递减
func (r *RedisCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return r.Increment(ctx, key, -delta)
}

// Expire 设置过期时间
func (r *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := r.client.Expire(ctx, r.buildKey(key), ttl).Err(); err != nil {
		logger.Errorf("Redis EXPIRE失败: %v", err)
		return cache.ErrCacheConnection.WithCause(err)
	}

	return nil
}

// TTL 获取剩余生存时间
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	ttl, err := r.client.TTL(ctx, r.buildKey(key)).Result()
	if err != nil {
		logger.Errorf("Redis TTL失败: %v", err)
		return 0, cache.ErrCacheConnection.WithCause(err)
	}

	return ttl, nil
}

// Keys 获取匹配的键
func (r *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	keys, err := r.client.Keys(ctx, r.buildKey(pattern)).Result()
	if err != nil {
		logger.Errorf("Redis KEYS失败: %v", err)
		return nil, cache.ErrCacheConnection.WithCause(err)
	}

	// 移除前缀
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = strings.TrimPrefix(key, r.prefix)
	}

	return result, nil
}

// DeleteByPattern 按模式删除
func (r *RedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	keys, err := r.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return r.MDelete(ctx, keys)
}

// Ping 连接测试
func (r *RedisCache) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := r.client.Ping(ctx).Err(); err != nil {
		return cache.ErrCacheConnection.WithCause(err)
	}

	return nil
}

// Close 关闭连接
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// TypedRedisCache 类型化Redis缓存
type TypedRedisCache[T any] struct {
	cache *RedisCache
}

// NewTypedRedisCache 创建类型化Redis缓存
func NewTypedRedisCache[T any](cache *RedisCache) *TypedRedisCache[T] {
	return &TypedRedisCache[T]{cache: cache}
}

// Get 获取类型化数据
func (t *TypedRedisCache[T]) Get(ctx context.Context, key string) (*T, error) {
	data, err := t.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, cache.ErrCacheSerialization.WithCause(err)
	}

	return &result, nil
}

// Set 设置类型化数据
func (t *TypedRedisCache[T]) Set(ctx context.Context, key string, value *T, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return cache.ErrCacheSerialization.WithCause(err)
	}

	return t.cache.Set(ctx, key, data, ttl)
}

// Delete 删除
func (t *TypedRedisCache[T]) Delete(ctx context.Context, key string) error {
	return t.cache.Delete(ctx, key)
}

// Exists 检查存在
func (t *TypedRedisCache[T]) Exists(ctx context.Context, key string) (bool, error) {
	return t.cache.Exists(ctx, key)
}

// MGet 批量获取类型化数据
func (t *TypedRedisCache[T]) MGet(ctx context.Context, keys []string) (map[string]*T, error) {
	data, err := t.cache.MGet(ctx, keys)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*T)
	for key, bytes := range data {
		var value T
		if err := json.Unmarshal(bytes, &value); err != nil {
			logger.Warnf("反序列化键 %s 失败: %v", key, err)
			continue
		}
		result[key] = &value
	}

	return result, nil
}

// MSet 批量设置类型化数据
func (t *TypedRedisCache[T]) MSet(ctx context.Context, items map[string]*T, ttl time.Duration) error {
	data := make(map[string][]byte)
	for key, value := range items {
		bytes, err := json.Marshal(value)
		if err != nil {
			return cache.ErrCacheSerialization.WithCause(err)
		}
		data[key] = bytes
	}

	return t.cache.MSet(ctx, data, ttl)
}
