package campaign

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

type redisCase struct {
	name   string
	ctx    context.Context
	addr   string
	expVal bool
	expErr bool
}

func TestClient(tester *testing.T) {
	// RedisClient
	tester.Run("NewRedisClient Test", func(test *testing.T) {
		// Definition
		testCases := []redisCase{
			{
				"Invalid address error",
				context.Background(),
				"localhost:678",
				false,
				true,
			},
			{
				"Correct connection",
				context.Background(),
				RedisUrl,
				true,
				false,
			},
		}
		// Execution
		for _, tc := range testCases {
			test.Run(tc.name, func(t *testing.T) {
				res, err := NewRedisClient(tc.ctx, tc.addr)
				noRes := res != nil
				noErr := err != nil
				assert.True(t, tc.expVal == noRes, "Got result: %v, Exp: %v", !noRes, tc.expVal)
				assert.True(t, tc.expErr == noErr, "Got err: %v, Exp: %v", !noErr, tc.expErr)
			})
		}
	})
}
