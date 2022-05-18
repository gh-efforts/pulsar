package subscriber

import "context"

type MsgLocker interface {
	Acquire(ctx context.Context, key string) (bool, error)
	Release(ctx context.Context, key string) bool
}
