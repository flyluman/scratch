package valkey

import (
	"fmt"

	"github.com/flyluman/scratch/internal/platform/config"
	"github.com/valkey-io/valkey-go"
)

type Client struct {
	valkey.Client
}

func NewClient(cfg config.Config) (*Client, error) {
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{cfg.ValkeyAddr},
		Username:    cfg.ValkeyUsername,
		Password:    cfg.ValkeyPassword,
		SelectDB:    cfg.ValkeyDB,
	})
	if err != nil {
		return nil, fmt.Errorf("valkey connect: %w", err)
	}
	return &Client{client}, nil
}
