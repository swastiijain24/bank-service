package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	redisClient *redis.Client
	ttl         time.Duration
	prefix      string
}

func NewRedisStore(addr string, ttl time.Duration) *RedisStore {
	return &RedisStore{
		redisClient: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
		ttl:    ttl,
		prefix: "idempotency:",
	}
}

func (s *RedisStore) GetRedisResponse(key string) ([]byte, error) {
	ctx := context.Background()

	data, err := s.redisClient.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err

	}
	return data, nil
}


func (s *RedisStore) SetRedisResponse(key string, data []byte) error {
	ctx := context.Background()
	ok, err := s.redisClient.SetArgs(ctx, key, "processing", redis.SetArgs{
		Mode: "NX",
		TTL:  5 * time.Minute,
	}).Result()
	if err != nil {
		return  fmt.Errorf("error setting to redis: %s" , err)
	}

	if ok == "OK" {
		return s.redisClient.Set(ctx, key, data, s.ttl).Err()
	} 
	return nil 
}
