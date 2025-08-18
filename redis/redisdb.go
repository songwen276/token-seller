package redisdb

import (
	"context"
	"fmt"
	"sync"
	"time"
	"token-seller/config"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	once        sync.Once
	mu          sync.RWMutex
)

func GetRedisClient() (*redis.Client, error) {
	mu.RLock()
	// 先检查实例是否存在，存在则直接返回
	if redisClient != nil {
		// 简单检查连接是否正常
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if _, err := redisClient.Ping(ctx).Result(); err == nil {
			mu.RUnlock()
			return redisClient, nil
		}
		// 连接异常，释放读锁，准备重新创建
		mu.RUnlock()
	} else {
		mu.RUnlock()
	}

	// 使用sync.Once确保只创建一次实例
	var err error
	once.Do(func() {
		mu.Lock()
		defer mu.Unlock()

		// 再次检查，避免在等待锁的过程中其他goroutine已经创建了实例
		if redisClient != nil {
			if _, err := redisClient.Ping(context.Background()).Result(); err == nil {
				return
			}
			// 连接异常，关闭旧连接
			_ = redisClient.Close()
		}

		// 创建新的Redis客户端
		redisClient = redis.NewClient(&redis.Options{
			Addr:         config.ConfigCache.Redis.Addr,
			Password:     config.ConfigCache.Redis.Password,
			DB:           config.ConfigCache.Redis.Db, // 使用默认数据库
			DialTimeout:  10 * time.Second,            // 增加连接超时时间
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			PoolSize:     10,
			MinIdleConns: 5,
		})

		// 测试连接
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err = redisClient.Ping(ctx).Result()
		if err != nil {
			redisClient = nil // 创建失败，重置为nil
		}
	})

	if err != nil {
		return nil, fmt.Errorf("创建Redis客户端失败: %v", err)
	}

	return redisClient, nil
}

// CloseRedisClient 关闭Redis连接
func CloseRedisClient() error {
	mu.Lock()
	defer mu.Unlock()

	if redisClient != nil {
		err := redisClient.Close()
		redisClient = nil
		return err
	}
	return nil
}
