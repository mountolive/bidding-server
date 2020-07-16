package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mountolive/bidding-server/campaign"
	"github.com/mountolive/bidding-server/search"
)

type response struct {
	Price float64 `json:"price"`
}

func main() {
	// Creating the service
	bidder := search.Bidder{}
	// Warmup - Periodic Loading Bid data in memory
	// This generates a channel that will be consumed by each req
	campaigns := campaign.GenerateRedisData(context.Background())

	fmt.Println("Starting server: localhost:8080")

	http.HandleFunc("/bid", func(w http.ResponseWriter, r *http.Request) {
		pubId, ok := r.URL.Query()["publisherid"]
		var maxPrice float64
		if ok && len(pubId) > 0 {
			pos, ok := r.URL.Query()["position"]
			if ok && len(pos) > 0 {
				maxPrice = processRequest(bidder, pos[0], pubId[0], campaigns)
			}
		}
		if maxPrice == 0 {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response{maxPrice})
		}
	})

	http.ListenAndServe(":8080", nil)
}

func processRequest(bidder search.Bidder, pos, pubId string, cmpgs <-chan []campaign.Campaign) float64 {
	// Asserting params - accepting only ints
	var position int
	var publisherId int
	var err error
	position, err = strconv.Atoi(pos)
	if err != nil {
		fmt.Println("Wrong value for value for position parameter, must be an int")
		return 0
	}
	publisherId, err = strconv.Atoi(pubId)
	// For simplicity, we'll cancel the lookup if the parameters are wrong
	// and just return a 204
	if err != nil {
		fmt.Println("Wrong value for value for publisherid parameter, must be an int")
		return 0
	}
	// This releases "the lock" for the GenerateRedisData to rebuild the data
	campaigns, ok := <-cmpgs
	// Means channel is closed for some reason
	if !ok {
		fmt.Println("Something went wrong when retrieving the data from redis")
		return 0
	}
	maxPrice := bidder.BestBid(position, publisherId, campaigns)
	return maxPrice
}
