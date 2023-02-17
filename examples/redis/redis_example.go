package main

import (
	"context"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	tcredis "github.com/tordf/testcontainers/redis"
)

func run() string {
	log.SetLevel(log.InfoLevel)
	// Start redis container
	redisCont, redisConf, err := tcredis.StartRedisContainer(context.Background(), tcredis.ContainerOptions{})
	if err != nil {
		log.Fatalf("Failed to start redis container: %v", err)
	}
	defer redisCont.Terminate(context.Background())

	// Connect to redis database
	db := redis.NewClient(&redis.Options{
		Addr:     redisConf.ConnectionURI(),
		Password: redisConf.Password,
		DB:       1,
	})

	// Set some data
	db.HSet("my-hash-key", "key1", "Hello ")
	db.HSet("my-hash-key", "key2", "World!")

	// Get the data back
	k1, _ := db.HGet("my-hash-key", "key1").Result() // "Hello "
	k2, _ := db.HGet("my-hash-key", "key2").Result() // "World!"

	return k1 + k2
}

func main() {
	log.Infof("Received %s from redis", run())
}
