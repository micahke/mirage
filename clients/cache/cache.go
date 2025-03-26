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
}
