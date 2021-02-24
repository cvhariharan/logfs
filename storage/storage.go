package storage

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type Object struct {
	Contents     []byte
	LastModified time.Time
}

type StorageBackend interface {
	PutContents(ctx context.Context, path string, content []byte) error
	GetContents(ctx context.Context, path string) ([]byte, error)
	ListAllByPrefix(ctx context.Context, prefix string) ([]interface{}, error)
	DeleteContents(ctx context.Context, path string) error
}

type RedisStorage struct {
	Client *redis.Client
}

func NewRedisStorage(addr, password string) *RedisStorage {
	return &RedisStorage{
		Client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       0,
		}),
	}
}

func (r *RedisStorage) PutContents(ctx context.Context, name string, content []byte) error {
	err := r.Client.Set(ctx, name, content, 0).Err()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (r *RedisStorage) GetContents(ctx context.Context, path string) ([]byte, error) {
	val, err := r.Client.Get(ctx, path).Bytes()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return val, nil
}
