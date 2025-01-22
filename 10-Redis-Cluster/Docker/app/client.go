package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

var (
	redisHost     = os.Getenv("REDIS_HOST")
	redisPort     = os.Getenv("REDIS_PORT")
	redisPassword = os.Getenv("REDIS_PASSWORD")
	ctx           = context.Background()
	rdb           *redis.Client
)

func init() {
	// Validate Redis configuration
	if redisHost == "" || redisPort == "" {
		log.Fatal("REDIS_HOST and REDIS_PORT must be set")
	}

	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword, // Use password from environment, if set
		DB:       0,             // Default DB
	})

	// Test Redis connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Info("Connected to Redis successfully")
}

func main() {
	router := httprouter.New()

	router.GET("/", incrementRedisKey)

	log.Info("Server is running on port 80")
	log.Fatal(http.ListenAndServe(":80", router))
}

func incrementRedisKey(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	var counter int

	// Get the current counter value from Redis
	val, err := rdb.Get(ctx, "counter").Result()
	if err == redis.Nil {
		// Key does not exist, initialize it
		counter = 1
		err = rdb.Set(ctx, "counter", counter, 0).Err()
		if err != nil {
			log.WithError(err).Error("Failed to initialize counter")
			http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		log.Info("Counter initialized to 1")
	} else if err != nil {
		log.WithError(err).Error("Failed to get counter from Redis")
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	} else {
		// Parse the counter value and increment it
		counter, err = strconv.Atoi(val)
		if err != nil {
			log.WithError(err).Error("Failed to parse counter value")
			http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		counter++
		err = rdb.Set(ctx, "counter", counter, 0).Err()
		if err != nil {
			log.WithError(err).Error("Failed to update counter in Redis")
			http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// Respond with the updated counter value
	writer.WriteHeader(http.StatusOK)
	fmt.Fprint(writer, counter)
	log.Infof("Counter updated: %d", counter)
}
