package database

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

func NewRedis(addr, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return rdb, nil
}

// NewRedisOptional 创建可选的 Redis 连接，如果连接失败只记录警告
func NewRedisOptional(addr, password string, db int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("⚠️  Redis connection failed: %v", err)
		log.Printf("⚠️  Some features will be disabled (verification codes, password reset)")
		return nil
	}

	log.Printf("✅ Redis connected successfully")
	return rdb
}
