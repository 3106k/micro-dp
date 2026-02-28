package queue

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

type ValkeyClient struct {
	rdb *redis.Client
}

func NewValkeyClient() (*ValkeyClient, error) {
	addr := os.Getenv("VALKEY_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("valkey ping: %w", err)
	}

	return &ValkeyClient{rdb: rdb}, nil
}

func (c *ValkeyClient) Client() *redis.Client {
	return c.rdb
}

func (c *ValkeyClient) Close() error {
	return c.rdb.Close()
}
