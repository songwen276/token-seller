package utils

import (
	"context"
	"encoding/json"
	"time"
	redisdb "token-seller/redis"
)

// 默认上下文（可根据需要替换为带超时的上下文）
var defaultCtx = context.Background()

// RedisSet 设置键值对（带过期时间）
func RedisSet(key string, value interface{}, expiration time.Duration) error {
	client, err := redisdb.GetRedisClient()
	if err != nil {
		return err
	}
	return client.Set(defaultCtx, key, value, expiration).Err()
}

// RedisGet 获取键值
func RedisGet(key string) (string, error) {
	client, err := redisdb.GetRedisClient()
	if err != nil {
		return "", err
	}
	return client.Get(defaultCtx, key).Result()
}

// RedisGetObject 获取键值并解析为对象
func RedisGetObject(key string, obj interface{}) error {
	val, err := RedisGet(key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), obj)
}

// RedisSetObject 存储对象（自动序列化为JSON）
func RedisSetObject(key string, obj interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return RedisSet(key, jsonData, expiration)
}

// RedisExists 判断键是否存在
func RedisExists(key string) (bool, error) {
	client, err := redisdb.GetRedisClient()
	if err != nil {
		return false, err
	}
	count, err := client.Exists(defaultCtx, key).Result()
	return count > 0, err
}

// RedisDelete 删除键
func RedisDelete(key string) error {
	client, err := redisdb.GetRedisClient()
	if err != nil {
		return err
	}
	return client.Del(defaultCtx, key).Err()
}

// RedisIncr 原子递增
func RedisIncr(key string) (int64, error) {
	client, err := redisdb.GetRedisClient()
	if err != nil {
		return 0, err
	}
	return client.Incr(defaultCtx, key).Result()
}

// RedisHSet 设置哈希表字段
func RedisHSet(key, field string, value interface{}) error {
	client, err := redisdb.GetRedisClient()
	if err != nil {
		return err
	}
	return client.HSet(defaultCtx, key, field, value).Err()
}

// RedisHGet 获取哈希表字段
func RedisHGet(key, field string) (string, error) {
	client, err := redisdb.GetRedisClient()
	if err != nil {
		return "", err
	}
	return client.HGet(defaultCtx, key, field).Result()
}

// RedisLock 获取分布式锁
func RedisLock(key string, value string, expiration time.Duration) (bool, error) {
	client, err := redisdb.GetRedisClient()
	if err != nil {
		return false, err
	}
	return client.SetNX(defaultCtx, key, value, expiration).Result()
}

// RedisUnlock 释放分布式锁
func RedisUnlock(key string) error {
	return RedisDelete(key)
}

// 其他常用方法（如列表、有序集合等）可参考上述格式添加
