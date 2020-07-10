package search

import (
	"context"
	"github.com/mountolive/bidding-server/campaign"
)

type Bid interface {
	FindBestBid(ctx context.Context, p, id int, cmpgs []campaign.Campaign, candidate SearchCandidate, lookup Search) float64
}

type Bidder struct{}

// Wrapper method that uses default implementations for search of best price
func (b Bidder) BestBid(ctx context.Context, pos, pubId int, cmpgs []campaign.Campaign) float64 {
	return b.FindBestBid(ctx, pos, pubId, cmpgs, candidateSearcher{}, searcher{})
}

// Find the best bid among all campaigns
// if not, maxPrice would be 0
// We ignore any error and keep traversing (But it should be handled)
func (b Bidder) FindBestBid(ctx context.Context, pos, pubId int, cmpgs []campaign.Campaign, candidate SearchCandidate, lookup Search) (maxPrice float64) {
	if len(cmpgs) < 1 {
		return
	}
	// Constraint vars
	var pubConst bool
	var posConst bool

	for _, cmpg := range cmpgs {
		// Check if it's among the publishers (this lookup should be
		// less expensive)
		pubConst, _ = candidate.Candidate(ctx, pubId, cmpg.Publishers, lookup)
		if pubConst {
			posConst, _ = candidate.Candidate(ctx, pos, cmpg.Positions, lookup)
			if posConst {
				if maxPrice <= cmpg.Cpm {
					maxPrice = cmpg.Cpm
				}
			}
		}
	}
	return
}
