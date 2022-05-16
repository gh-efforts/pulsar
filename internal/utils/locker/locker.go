package locker

import (
	"context"

	"github.com/bitrainforest/filmeta-hic/core/store"

	"github.com/bitrainforest/filmeta-hic/core/store/redisx"
)

const (
	// DefaultLockExpire seconds
	DefaultLockExpire uint32 = 3600
)

// NewRedisLock  lockExpire is the  lock expire time in seconds
func NewRedisLock(ctx context.Context, key string, lockExpire uint32) *redisx.RedisLock {
	expire := DefaultLockExpire
	if lockExpire > 0 {
		expire = lockExpire
	}
	return redisx.NewRedisLock(store.GetRedisClient(ctx), key, redisx.SetLockExpire(expire))
}
