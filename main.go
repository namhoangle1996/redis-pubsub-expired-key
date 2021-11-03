package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)
//Docs https://medium.com/nerd-for-tech/redis-getting-notified-when-a-key-is-expired-or-changed-ca3e1f1c7f0a
func main() {
	// connect to redis
	redisDB := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	_, err := redisDB.Do(context.Background(), "CONFIG", "SET", "notify-keyspace-events", "KEA").Result() // this is telling redis to publish events since it's off by default.
	if err != nil {
		fmt.Printf("unable to set keyspace events %v", err.Error())
		os.Exit(1)
	}


	//Having 4 Channels :
	//	pmessage
	// 	__key*__:*
	//	__keyspace@0__:mykey
	//	set


	pubsub := redisDB.PSubscribe(context.Background(), "__keyevent@0__:expired")
	// this is telling redis to subscribe to events published in the keyevent channel, specifically for expired events
	// this is just to show publishing events and catching the expired events in the same codebase.
	wg := &sync.WaitGroup{}
	wg.Add(2) // two goroutines are spawned

	go func(redis.PubSub) {
		exitLoopCounter := 0
		for { // infinite loop
			// this listens in the background for messages.
			message, err := pubsub.ReceiveMessage(context.Background())
			exitLoopCounter++
			if err != nil {
				fmt.Printf("error message - %v", err.Error())
				break
			}
			fmt.Printf("MessagePayload.recieved %v  \n", message.Payload)

			if exitLoopCounter >= 10 {
				wg.Done()
			}
		}
	}(*pubsub)

	go func(redis.Client, *sync.WaitGroup) {
		for i := 0; i <= 10; i++ {
			dynamicKey := fmt.Sprintf("order/%v", i)
			redisDB.Set(context.Background(), dynamicKey, "test_value"+ fmt.Sprintf("%v",i) , time.Second*time.Duration(i*3) )
		}
		wg.Done()
	}(*redisDB, wg)

	wg.Wait()
	fmt.Println("exiting program")
}