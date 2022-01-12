package dsf

import (
	"context"
	"time"
)

type RedisClient interface {
	Context(ctx context.Context) RedisClient

	Nil() interface{}
	Get(key string) ([]byte, error)
	Set(key string, val []byte, exp time.Duration) error
	Eval(script string, keys []string, args ...interface{}) (interface{}, error)
}
