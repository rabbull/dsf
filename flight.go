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
			client = client.Context(opt.ctx)
		}

		logger := opt.logger
		if logger != nil {
			logger = logger.Context(opt.ctx)
		}

		return &groupImpl{
			ctx:       opt.ctx,
			logger:    &loggerWrapper{logger: logger},
			client:    client,
			namespace: opt.namespace,

			lockExp:  opt.lockExp,
			dataExp:  opt.dataExp,
			keepLock: opt.keepLock,

			resultWaitTime: opt.resultWaitTime,
			makeInterval:   opt.makeInterval,
		}
	}
	return nil
}

type groupImpl struct {
	ctx       context.Context
	logger    *loggerWrapper
	client    RedisClient
	namespace string

	lockExp  time.Duration
	dataExp  time.Duration
	keepLock bool

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
	f.logger.D("threadID=%v", threadID)

	lockKey := f.makeLockKey(jobKey)
	f.logger.D("lockKey=%v", lockKey)

	scriptResult, err := f.client.Eval(_SetNXOrGet, []string{
		lockKey,
	}, []byte(threadID), f.lockExp.Seconds())
	if err != nil {
		return nil, false, err
	}
	isMaster := scriptResult.([]interface{})[0].(int64) == 1
	masterID := scriptResult.([]interface{})[1].(string)
	f.logger.D("isMaster=%v, masterID=%v", isMaster, masterID)

	if !isMaster {
		f.logger.I("lock missed: lockKey=%v, masterID=%v", lockKey, masterID)
		sharedResult, err := f.waitForResult(masterID)
		return sharedResult, true, err
	}

	f.logger.I("acquired lock: lockKey=%v, threadID=%v", lockKey, threadID)

	jobResult := job()
	f.logger.D("jobResult=%v, str=%v", jobResult, string(jobResult))

	dataKey := f.makeDataKey(threadID)
	f.logger.D("dataKey=%v", dataKey)

	if err = f.client.Set(dataKey, jobResult, f.dataExp); err != nil {
		return jobResult, false, nil
	}

	if !f.keepLock {
		f.logger.I("releasing lock: lockKey=%v", lockKey)
		releaseRes, err := f.client.Eval(_DelIfEq, []string{
			lockKey,
		}, []byte(threadID))
		f.logger.I("releaseRes=%v", releaseRes)

		if err != nil {
			return jobResult, false, err
		}
	}

	return jobResult, false, nil
}

func (f *groupImpl) waitForResult(masterID string) ([]byte, error) {
	dataKey := f.makeDataKey(masterID)
	deadline := time.Now().Add(f.resultWaitTime)
	for retryTimes := 0; time.Now().Before(deadline); retryTimes += 1 {
		f.logger.D("retryTimes=%v", retryTimes)
		sharedResult, ok, err := f.tryFetchResult(dataKey)
		if err != nil {
			return nil, err
		}
		if ok {
			f.logger.D("sharedResult=%v, str=%v",
				sharedResult, string(sharedResult))
			return sharedResult, nil
		}

		f.logger.I(
			"result not found: dataKey=%v, retryTimes=%v, timeLeft=%v",
			dataKey, retryTimes, time.Until(deadline),
		)
		interval := f.makeInterval(retryTimes)
		f.logger.D("interval=%vs", interval.Seconds())
		time.Sleep(interval)
	}
	f.logger.E("result not found after %vs", f.resultWaitTime.Seconds())
	return nil, errors.New("result timeout")
}

func (f *groupImpl) tryFetchResult(dataKey string) ([]byte, bool, error) {
	sharedResult, err := f.client.Get(dataKey)
	f.logger.D("sharedResult=%v, err=%v", sharedResult, err)
	if err == f.client.Nil() {
		return nil, false, nil
	} else if err != nil && err != f.client.Nil() {
		return nil, false, err
	}
	return sharedResult, true, nil
}
