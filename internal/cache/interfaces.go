package cache

import (
	"context"
	"time"
)

// Cache 缓存接口
type Cache interface {
	// 基础操作
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	
	// 批量操作
	MGet(ctx context.Context, keys []string) (map[string][]byte, error)
	MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error
	MDelete(ctx context.Context, keys []string) error
	
	// 高级操作
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	Decrement(ctx context.Context, key string, delta int64) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	
	// 模式匹配
	Keys(ctx context.Context, pattern string) ([]string, error)
	DeleteByPattern(ctx context.Context, pattern string) error
	
	// 连接管理
	Ping(ctx context.Context) error
	Close() error
}

// TypedCache 类型化缓存接口
type TypedCache[T any] interface {
	Get(ctx context.Context, key string) (*T, error)
	Set(ctx context.Context, key string, value *T, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	MGet(ctx context.Context, keys []string) (map[string]*T, error)
	MSet(ctx context.Context, items map[string]*T, ttl time.Duration) error
}

// CacheManager 缓存管理器接口
type CacheManager interface {
	// 获取不同类型的缓存
	GetCache(name string) Cache
	
	// 缓存管理
	CreateCache(name string, config CacheConfig) error
	RemoveCache(name string) error
	
	// 统计信息
	GetStats(ctx context.Context) (map[string]CacheStats, error)
	
	// 健康检查
	HealthCheck(ctx context.Context) error
	
	// 关闭
	Close() error
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Type        string        `json:"type"`         // redis, memory, hybrid
	TTL         time.Duration `json:"ttl"`          // 默认TTL
	MaxSize     int64         `json:"max_size"`     // 最大大小
	MaxItems    int           `json:"max_items"`    // 最大条目数
	Compression bool          `json:"compression"`  // 是否压缩
	Serializer  string        `json:"serializer"`   // json, msgpack, gob
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	HitRate     float64 `json:"hit_rate"`
	Size        int64   `json:"size"`
	Items       int     `json:"items"`
	Evictions   int64   `json:"evictions"`
	Connections int     `json:"connections"`
}

// 预定义错误
var (
	ErrCacheNotFound    = NewCacheError("CACHE_NOT_FOUND", "缓存未找到")
	ErrCacheKeyNotFound = NewCacheError("KEY_NOT_FOUND", "键未找到")
	ErrCacheExpired     = NewCacheError("CACHE_EXPIRED", "缓存已过期")
	ErrCacheTimeout     = NewCacheError("CACHE_TIMEOUT", "缓存操作超时")
	ErrCacheConnection  = NewCacheError("CONNECTION_ERROR", "缓存连接错误")
	ErrCacheSerialization = NewCacheError("SERIALIZATION_ERROR", "序列化错误")
)

// CacheError 缓存错误
type CacheError struct {
	Code    string
	Message string
	Cause   error
}

func (e *CacheError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *CacheError) Unwrap() error {
	return e.Cause
}

// NewCacheError 创建缓存错误
func NewCacheError(code, message string) *CacheError {
	return &CacheError{
		Code:    code,
		Message: message,
	}
}

// WithCause 添加原因
func (e *CacheError) WithCause(cause error) *CacheError {
	return &CacheError{
		Code:    e.Code,
		Message: e.Message,
		Cause:   cause,
	}
}
