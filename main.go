package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/mountolive/bidding-server/campaign"
	"github.com/mountolive/bidding-server/examples"
	"github.com/mountolive/bidding-server/search"
)

func main() {
	// Creating the service
	bidder := search.Bidder{}
	ctx := context.Background()
	// Connecting to redis
	rdb, err := campaign.NewRedisClient(ctx, campaign.RedisUrl)
	// Defering closing
	defer rdb.Close()
	if err != nil {
		log.Fatalf("An error occurred when connecting to redis: %v", err)
	}

	fmt.Println("Starting server: localhost:8080")
	// Example server for redis-direct connection
	examples.StartWithRedisServer(ctx, rdb, bidder)

	http.ListenAndServe(":8080", nil)
}
