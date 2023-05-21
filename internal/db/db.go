package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v9"
)

func New(addr, password string, ttl time.Duration) *DB {
	return &DB{
		Client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       0, // use default DB
		}),
		ttl: ttl,
	}

}

type DB struct {
	Client *redis.Client
	ttl    time.Duration
}

func (db *DB) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = db.Client.Set(ctx, key, data, db.ttl).Result()

	return err
}

func (db *DB) Get(ctx context.Context, key string, value interface{}) error {
	data, err := db.Client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &value)
}
