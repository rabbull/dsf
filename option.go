package dsf

import (
	"context"
	"time"
)

type Option interface{}

type OptionBuilder interface {
	Build() Option

	WithContext(ctx context.Context) OptionBuilder
	WithRedisClient(client RedisClient) OptionBuilder
	WithNamespace(namespace string) OptionBuilder

	WithLockExpiration(exp time.Duration) OptionBuilder
	WithDataExpiration(exp time.Duration) OptionBuilder
	KeepLock(flag bool) OptionBuilder
	WithWaitTime(duration time.Duration) OptionBuilder
	WithInterval(makeInterval func(retryTimes int) time.Duration) OptionBuilder
}

func NewOptionBuilder() OptionBuilder {
	return &optionBuilderImpl{
		ctx:            context.Background(),
		namespace:      "default",
		lockExp:        time.Second,
		dataExp:        2 * time.Second,
		resultWaitTime: 2 * time.Second,

		makeInterval: func(retryTimes int) time.Duration {
			return 50 * time.Millisecond
		},
	}
}

type optionBuilderImpl struct {
	ctx            context.Context
	client         RedisClient
	namespace      string
	lockExp        time.Duration
	dataExp        time.Duration
	keepLock       bool
	resultWaitTime time.Duration
	makeInterval   func(retryTimes int) time.Duration
}

func (b *optionBuilderImpl) Build() Option {
	if b.client == nil ||
		b.lockExp <= 0 ||
		b.dataExp <= 0 ||
		b.resultWaitTime <= 0 {
		return nil
	}
	cpy := *b
	return &cpy
}

func (b *optionBuilderImpl) WithContext(ctx context.Context) OptionBuilder {
	b.ctx = ctx
	return b
}

func (b *optionBuilderImpl) WithRedisClient(client RedisClient) OptionBuilder {
	b.client = client
	return b
}

func (b *optionBuilderImpl) WithNamespace(namespace string) OptionBuilder {
	b.namespace = namespace
	return b
}

func (b *optionBuilderImpl) WithLockExpiration(t time.Duration) OptionBuilder {
	b.lockExp = t
	return b
}

func (b *optionBuilderImpl) WithDataExpiration(t time.Duration) OptionBuilder {
	b.dataExp = t
	return b
}

func (b *optionBuilderImpl) KeepLock(flag bool) OptionBuilder {
	b.keepLock = flag
	return b
}

func (b *optionBuilderImpl) WithWaitTime(duration time.Duration) OptionBuilder {
	b.resultWaitTime = duration
	return b
}

func (b *optionBuilderImpl) WithInterval(
	makeInterval func(retryTimes int) time.Duration,
) OptionBuilder {
	b.makeInterval = makeInterval
	return b
}
