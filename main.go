package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	// Warmup - Loading Bid data in memory
	campaigns, err := campaign.LoadDefaultCampaigns()
	if err != nil {
		// For simplicity, we'll stop the server if an error occurred in the warmup
		log.Fatalf("An error occurred during warmup of campaign data: %v", err)
	}

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
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response{maxPrice})
		}
	})

	http.ListenAndServe(":8080", nil)
}

func processRequest(bidder search.Bidder, pos, pubId string, cmpgs []campaign.Campaign) float64 {
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
	// One context per request
	ctx := context.Background()
	maxPrice := bidder.BestBid(ctx, position, publisherId, cmpgs)
	return maxPrice
}
