package dsf

import (
	"context"
	"encoding/ascii85"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Job func() []byte

type Group interface {
	Do(jobKey string, job Job) ([]byte, bool, error)
}

func New(opt Option) Group {
	if opt == nil {
		return nil
	}

	switch opt := opt.(type) {
	case *optionBuilderImpl:
		client := opt.client
		if opt.ctx != nil {
			client = opt.client.Context(opt.ctx)
		}

		return &groupImpl{
			ctx:       opt.ctx,
			client:    client,
			namespace: opt.namespace,

			lockExp: opt.lockExp,
			dataExp: opt.dataExp,

			resultWaitTime: opt.resultWaitTime,
			makeInterval:   opt.makeInterval,
		}
	}
	return nil
}

type groupImpl struct {
	ctx       context.Context
	client    RedisClient
	namespace string

	lockExp time.Duration
	dataExp time.Duration

	resultWaitTime time.Duration
	makeInterval   func(retryTime int) time.Duration
}

func (f *groupImpl) makeLockKey(key string) string {
	return fmt.Sprintf("dsf://%v/?k=%v", f.namespace, key)
}

func (f *groupImpl) makeDataKey(threadID string) string {
	return fmt.Sprintf("dsf://%v/?t=%v", f.namespace, threadID)
}

func (f *groupImpl) genID() (string, error) {
	rawThreadID, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	buffer := make([]byte, ascii85.MaxEncodedLen(len(rawThreadID)))
	_ = ascii85.Encode(buffer, rawThreadID[:])
	return string(buffer), nil
}

func (f *groupImpl) Do(jobKey string, job Job) ([]byte, bool, error) {

	threadID, err := f.genID()
	if err != nil {
		return nil, false, err
	}

	lockKey := f.makeLockKey(jobKey)
	scriptResult, err := f.client.Eval(setNXOrGet, []string{
		lockKey,
	}, []byte(threadID), f.lockExp.Seconds())
	if err != nil {
		return nil, false, err
	}
	isMaster := scriptResult.([]interface{})[0].(int64) == 1
	masterID := scriptResult.([]interface{})[1].(string)

	if !isMaster {
		sharedResult, err := f.waitForResult(masterID)
		return sharedResult, true, err
	}

	jobResult := job()
	dataKey := f.makeDataKey(threadID)
	if err = f.client.Set(dataKey, jobResult, f.dataExp); err != nil {
		return jobResult, false, nil
	}
	if err := f.client.Del(lockKey); err != nil {
		return jobResult, false, nil
	}

	return jobResult, false, nil
}

func (f *groupImpl) waitForResult(masterID string) ([]byte, error) {
	deadline := time.Now().Add(f.resultWaitTime)
	for retryTimes := 0; time.Now().Before(deadline); retryTimes += 1 {
		dataKey := f.makeDataKey(masterID)
		res, err := f.client.Get(dataKey)
		if err == f.client.Nil() {
			time.Sleep(f.makeInterval(retryTimes))
			continue
		} else if err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, errors.New("result timeout")
}
