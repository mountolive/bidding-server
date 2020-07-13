package search

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/mountolive/bidding-server/campaign"
)

type Bid interface {
	FindBestBidInMem(p, id int, cmpgs []campaign.Campaign, candidate SearchCandidate) float64
	FindBestBidRedis(ctx context.Context, rdb *redis.Client, pos, pubId int, candidate SearchCandidate) float64
}

type Bidder struct{}

// Wrapper method that uses default implementations for search of best price
func (b Bidder) BestBidInMem(pos, pubId int, cmpgs []campaign.Campaign) float64 {
	return b.FindBestBidInMem(pos, pubId, cmpgs, candidateSearcher{})
}

func (b Bidder) BestBidRedis(ctx context.Context, rdb *redis.Client, pos, pubId int) float64 {
	return b.FindBestBidRedis(ctx, rdb, pos, pubId, candidateSearcher{})
}

// Find the best bid among all campaigns (directly from redis)
// if not, maxPrice would be 0
// Ignoring any error and keep traversing (But it should be handled)
func (b Bidder) FindBestBidRedis(ctx context.Context, rdb *redis.Client, pos, pubId int, candidate SearchCandidate) (maxPrice float64) {

	campaigns := campaign.TraverseCampaigns(ctx, rdb)

	for cmpg := range campaigns {
		checkCampaign(candidate, cmpg, &maxPrice, pos, pubId)
	}
	return
}

// Find the best bid among all campaigns (in memory)
// if not, maxPrice would be 0
// We ignore any error and keep traversing (But it should be handled)
func (b Bidder) FindBestBidInMem(pos, pubId int, cmpgs []campaign.Campaign, candidate SearchCandidate) (maxPrice float64) {
	if len(cmpgs) < 1 {
		return
	}

	for _, cmpg := range cmpgs {
		checkCampaign(candidate, cmpg, &maxPrice, pos, pubId)
	}
	return
}

func checkCampaign(candidate SearchCandidate, cmpg campaign.Campaign, currentPrice *float64, pos, pubId int) {
	pubConst, _ := candidate.Candidate(pubId, cmpg.Publishers)
	if pubConst {
		posConst, _ := candidate.Candidate(pos, cmpg.Positions)
		if posConst {
			if *currentPrice <= cmpg.Cpm {
				*currentPrice = cmpg.Cpm
			}
		}
	}
}
