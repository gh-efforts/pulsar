package cache

import (
	"context"

	"github.com/bitrainforest/filmeta-hic/core/store"

	"github.com/go-redis/redis/v8"
)

var (
	_ AddressMark = (*AddressMarkCache)(nil)
)

type AddressMark interface {
	ExistAddress(ctx context.Context, address string) bool
	MarkAddress(ctx context.Context, address string) bool
}

type AddressMarkCache struct {
	store *redis.Client
}

func NewAddressMark(ctx context.Context) AddressMark {
	return &AddressMarkCache{
		store: store.GetRedisClient(ctx),
	}
}

func (u *AddressMarkCache) MarkAddress(ctx context.Context, address string) bool {
	intCmd := u.store.SetBit(ctx, address, 1, 1)
	return intCmd.Val() == 0
}

func (u *AddressMarkCache) ExistAddress(ctx context.Context, address string) bool {
	intCmd := u.store.GetBit(ctx, address, 1)
	res, err := intCmd.Result()
	if err != nil {
		return false
	}
	return res == 1
}
