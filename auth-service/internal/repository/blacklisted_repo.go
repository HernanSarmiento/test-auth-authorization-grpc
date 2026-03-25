package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type BlackListedToken interface {
	Save(ctx context.Context, jti string, expiration time.Duration) error
	Exists(ctx context.Context, jti string) (bool, error)
}

type RedisBlackListRepo struct {
	Client *redis.Client
	prefix string
}

func NewRedisBlackListRepo(client *redis.Client) BlackListedToken {
	return &RedisBlackListRepo{
		Client: client,
		prefix: "blacklist",
	}
}

func (r *RedisBlackListRepo) Save(ctx context.Context, jti string, expiration time.Duration) error {
	key := fmt.Sprintf("%s:%s", r.prefix, jti)

	err := r.Client.Set(ctx, key, "1", expiration).Err()
	if err != nil {
		return fmt.Errorf("Error: while adding token to the blacklist %w", err)
	}
	return nil
}

func (r *RedisBlackListRepo) Exists(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("%s%s", r.prefix, jti)

	val, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("Error: couldn't verify if token exists in black list %w", err)
	}
	return val > 0, nil
}
