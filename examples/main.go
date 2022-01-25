package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rabbull/dsf"
)

const NumWorkers = 16

func main() {
	sigDone := sync.WaitGroup{}
	sigDone.Add(NumWorkers)
	for i := 0; i < NumWorkers; i++ {
		time.Sleep(100 * time.Millisecond)
		go worker(i, &sigDone)
	}
	sigDone.Wait()
}

var job dsf.Job = func() []byte {
	println("job invoked")
	time.Sleep(time.Second * 2)
	bean := struct {
		Foo string  `json:"foo"`
		Bar float64 `json:"bar"`
	}{
		Foo: "fly!",
		Bar: math.Pi,
	}
	buf, err := json.Marshal(bean)
	if err != nil {
		panic(err)
	}
	println("job returned")
	return buf
}

func worker(workerID int, sigDone *sync.WaitGroup) {
	defer sigDone.Done()

	client := redis.NewClient(&redis.Options{
		Addr:        REDIS_ADDR,
		Username:    "default",
		Password:    REDIS_PASS,
		DialTimeout: 5 * time.Second,
	})
	if client == nil {
		println("client is nil")
		return
	}

	redisClient := &GoRedisClient{
		Ctx:    context.Background(),
		Client: client,
	}

	group := dsf.New(
		dsf.NewOptionBuilder().
			WithContext(context.Background()).
			WithRedisClient(redisClient).
			WithLockExpiration(5 * time.Second).
			WithDataExpiration(10 * time.Second).
			KeepLock(false).
			WithWaitTime(10 * time.Second).
			WithInterval(func(retryTimes int) time.Duration {
				return 0
			}).
			Build(),
	)
	if group == nil {
		println("group is nil")
		return
	}

	timeStart := time.Now()
	result, shared, err := group.Do("foobar", job)
	if err != nil {
		panic(err)
	}
	fmt.Printf("worker %d got result: shared=%v, result=%v, timeCost=%vs\n",
		workerID, shared, string(result), time.Since(timeStart).Seconds())
}
