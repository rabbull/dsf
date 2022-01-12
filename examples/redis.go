package main

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rabbull/dsf"
)

type GoRedisClient struct {
	Ctx    context.Context
	Client *redis.Client
}

func (c *GoRedisClient) Context(ctx context.Context) dsf.RedisClient {
	c.Ctx = ctx
	return c
}

func (c *GoRedisClient) Nil() interface{} {
	return redis.Nil
}

func (c *GoRedisClient) Get(key string) ([]byte, error) {
	value, err := c.Client.Get(c.Ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return []byte(value), nil
}

func (c *GoRedisClient) Set(key string, value []byte, exp time.Duration) error {
	return c.Client.Set(c.Ctx, key, string(value), exp).Err()
}

func (c *GoRedisClient) Del(key string) error {
	return c.Client.Del(c.Ctx, key).Err()
}

func (c *GoRedisClient) Eval(script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.Client.Eval(c.Ctx, script, keys, args).Result()
}
