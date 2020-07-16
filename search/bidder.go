package search

import (
	"fmt"
	"github.com/mountolive/bidding-server/campaign"
)

type Bid interface {
	FindBestBid(p, id int, cmpgs []campaign.Campaign, candidate SearchCandidate) float64
}

type Bidder struct{}

// Wrapper method that uses default implementations for search of best price
func (b Bidder) BestBid(pos, pubId int, cmpgs []campaign.Campaign) float64 {
	return b.FindBestBid(pos, pubId, cmpgs, candidateSearcher{})
}

// Find the best bid among all campaigns
// if not, maxPrice would be 0
// We ignore any error and keep traversing (But it should be handled)
func (b Bidder) FindBestBid(pos, pubId int, cmpgs []campaign.Campaign, candidate SearchCandidate) (maxPrice float64) {
	if len(cmpgs) == 0 {
		return
	}

	for _, cmpg := range cmpgs {
		// Checking first if the price matches constraint
		if maxPrice >= cmpg.Cpm {
			continue
		}
		// Check if it's among the publishers (this lookup should be
		// less expensive)
		pubConst, err := candidate.Candidate(pubId, cmpg.Publishers)
		if err != nil {
			// Just printing the error
			fmt.Printf("An error ocurred while trying to check publishers: %v, %v \n", cmpg.Id, err)
			// Skipping...
			continue
		}
		if pubConst {
			posConst, err := candidate.Candidate(pos, cmpg.Positions)
			if err != nil {
				// Just printing the error
				fmt.Printf("An error ocurred while trying to check positions: %v, %v \n", cmpg.Id, err)
				// Skipping
				continue
			}
			if posConst {
				maxPrice = cmpg.Cpm
			}
		}
	}
	return
}
