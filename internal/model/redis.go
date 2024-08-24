package models

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8" // Ensure you import the Redis package
	"go.mongodb.org/mongo-driver/bson"
)

type RedisHandler struct {
	client *redis.Client
	ctx    context.Context
}

var (
	redisHandler *RedisHandler
)

// NewRedisHandler initializes a new RedisHandler
func NewRedisHandler(addr string, password string, db int) *RedisHandler {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisHandler{
		client: rdb,
		ctx:    context.Background(),
	}
}

// SetUser stores the user data in Redis
func (r *RedisHandler) SetObject(key string, obj any, expiration time.Duration) error {
	objectJSON, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return r.client.Set(r.ctx, key, objectJSON, expiration).Err()
}

// GetObject retrieves the user data from Redis and always returns it as bson.M
func (r *RedisHandler) GetObject(key string) (bson.M, error) {
	// Get the raw JSON data from Redis
	objectData, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	fmt.Println("Raw JSON from Redis:", len(objectData))

	// Attempt to unmarshal into an interface to check the type
	var temp interface{}
	if err := json.Unmarshal([]byte(objectData), &temp); err != nil {
		return nil, err
	}

	result := bson.M{}

	switch v := temp.(type) {
	case []interface{}:
		// Data is an array of objects, store it under a key in bson.M
		result["data"] = v
	case map[string]interface{}:
		// Data is a single object, just convert it to bson.M
		result = v
	default:
		return nil, fmt.Errorf("unexpected data type in Redis")
	}

	fmt.Println("Converted to bson.M:", result)
	return result, nil
}

// GetUser retrieves the user data from Redis
func (r *RedisHandler) DeleteObject(key string) (*int64, error) {
	countDel, err := r.client.Del(r.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	return &countDel, nil
}

// GetRedis returns the initialized RedisHandler instance, creating it if necessary
func GetRedis() *RedisHandler {
	if redisHandler == nil {
		redisHandler = NewRedisHandler(os.Getenv("REDIS_URI"), "", 0)
	}
	return redisHandler
}
