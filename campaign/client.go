package campaign

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

// Wraps a redis client to test wether it has proper
// connection to it
func NewRedisClient(ctx context.Context, addr string) (*redis.Client, error) {
	// ReadTimeout is 5 secs by default (3 for Dial)
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	pong, err := rdb.Ping(ctx).Result()
	// Redis fails to respond
	if err != nil {
		return nil, err
	}
	fmt.Printf("Connected tor redis %s \n", pong)
	return rdb, nil
}
