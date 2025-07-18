package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(context.Context, string, interface{}) error
	GetMany(context.Context, []string, interface{}) error
	Set(context.Context, string, interface{}, time.Duration) error
	Delete(context.Context, string) error

	ScanKeys(context.Context, string) ([]string, error)

	Incr(context.Context, string) error
	IncrBy(context.Context, string, int64) (int64, error)
	Decr(context.Context, string) error
	DecrBy(context.Context, string, int64) (int64, error)
}
