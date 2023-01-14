package db

import (
	"context"
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
	_, err := db.Client.Conn().Set(ctx, key, value, db.ttl).Result()

	return err
}

func (db *DB) Get(ctx context.Context, key string) (interface{}, error) {
	return db.Client.Get(ctx, key).Result()
}
