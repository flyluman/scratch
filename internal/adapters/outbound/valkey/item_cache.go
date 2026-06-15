package valkey

import (
	"context"
	"fmt"
	"time"
)

type ItemCache struct {
	client *Client
}

func NewItemCache(client *Client) *ItemCache {
	return &ItemCache{client: client}
}

func (c *ItemCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Do(ctx, c.client.B().Get().Key(key).Build()).ToString()
	if err != nil {
		return "", fmt.Errorf("cache get: %w", err)
	}
	return val, nil
}

func (c *ItemCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Do(ctx, c.client.B().Set().Key(key).Value(value).Ex(ttl).Build()).Error()
}

func (c *ItemCache) Del(ctx context.Context, key string) error {
	return c.client.Do(ctx, c.client.B().Del().Key(key).Build()).Error()
}

func (c *ItemCache) Ping(ctx context.Context) error {
	return c.client.Do(ctx, c.client.B().Ping().Build()).Error()
}
