package main

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/go-redis/redis/v8"
	"github.com/rabbull/dsf"
)

const NumWorkers = 4

func main() {
	sigDone := sync.WaitGroup{}
	sigDone.Add(NumWorkers)
	sigTakeOff := sync.WaitGroup{}
	sigTakeOff.Add(1)
	for i := 0; i < NumWorkers; i++ {
		go worker(i, &sigDone, &sigTakeOff)
	}
	time.Sleep(time.Second)
	fmt.Println("ready to go")
	sigTakeOff.Done()
	sigDone.Wait()
}

var job dsf.Job = func() []byte {
	println("job invoked")
	time.Sleep(time.Millisecond * 500)
	bean := struct {
		Foo string  `json:"foo"`
		Bar float64 `json:"bar"`
	}{
		Foo: "fly!",
		Bar: math.Pi,
	}
	buf, err := sonic.Marshal(bean)
	if err != nil {
		panic(err)
	}
	return buf
}

func worker(workerID int, sigDone, sigTakeOff *sync.WaitGroup) {
	defer sigDone.Done()

	client := redis.NewClient(&redis.Options{
		Addr:     REDIS_ADDR,
		Password: REDIS_PASS,
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
			WithLockExpiration(time.Second).
			WithDataExpiration(time.Second).
			WithWaitTime(time.Second).
			WithInterval(func(retryTimes int) time.Duration {
				return 50 * time.Millisecond
			}).
			Build(),
	)
	if group == nil {
		println("group is nil")
		return
	}

	fmt.Printf("worker %d waiting\n", workerID)
	sigTakeOff.Wait()
	fmt.Printf("worker %d taking off\n", workerID)
	timeStart := time.Now()
	result, shared, err := group.Do("foobar", job)
	if err != nil {
		panic(err)
	}
	fmt.Printf("worker %d got result: shared=%v, result=%v, timeCost=%vms\n",
		workerID, shared, string(result), time.Now().Sub(timeStart).Milliseconds())
}
