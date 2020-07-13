package examples

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v8"

	"github.com/mountolive/bidding-server/search"
)

type Response struct {
	Price float64 `json:"price"`
}

func StartWithRedisServer(ctx context.Context, rdb *redis.Client, bidder search.Bidder) {
	http.HandleFunc("/bid", func(w http.ResponseWriter, r *http.Request) {
		pubId, ok := r.URL.Query()["publisherid"]
		var maxPrice float64
		if ok && len(pubId) > 0 {
			pos, ok := r.URL.Query()["position"]
			if ok && len(pos) > 0 {
				maxPrice = processRequestRedis(ctx, rdb, bidder, pos[0], pubId[0])
			}
		}
		if maxPrice == 0 {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Response{maxPrice})
		}
	})

}

func processRequestRedis(ctx context.Context, rdb *redis.Client, bidder search.Bidder, pos, pubId string) float64 {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// Asserting params - accepting only ints
	var position int
	var publisherId int
	var err error
	position, err = strconv.Atoi(pos)
	publisherId, err = strconv.Atoi(pubId)

	// For simplicity, we'll cancel the lookup if the parameters are wrong
	// and just return a 204
	if err != nil {
		return 0
	}
	maxPrice := bidder.BestBidRedis(ctx, rdb, position, publisherId)
	return maxPrice
}
