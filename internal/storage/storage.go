package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v9"
)

func New(addr, password string, ttl time.Duration) *Storage {
	return &Storage{
		Client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       0, // use default DB
		}),
		ttl: ttl,
	}

}

type Storage struct {
	Client *redis.Client
	ttl    time.Duration
}

func (s *Storage) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.Client.Set(ctx, key, data, s.ttl).Err()
}

func (s *Storage) Get(ctx context.Context, key string, value interface{}) error {
	data, err := s.Client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil
		}

		return err
	}

	return json.Unmarshal(data, &value)
}

func (s *Storage) Del(ctx context.Context, key string) error {
	return s.Client.Del(ctx, key).Err()
}
