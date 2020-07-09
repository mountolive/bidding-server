package search

import (
	"github.com/mountolive/bidding-server/campaign"
	"github.com/stretchr/testify/assert"
	"testing"
)

type bidderCase struct {
	name      string
	searcher  search
	candidate searchCandidate
	pubId     int
	pos       int
	data      []campaign.Campaign
	expVal    bool
}

type mockSearchCandidate struct{}

func (m mockSearchCandidate) Candidate(p int, data interface{}, srchr search) (bool, error) {
	return true, nil
}

// Base tests
func TestBidder(tester *testing.T) {
	pubId := 1
	pos := 300
	bidder := Bidder{}
	baseSearcher := searcher{}
	candSearcherMock := mockSearchCandidate{}
	properSearcher := candidateSearcher{}
	exampleCampaigns := []campaign.Campaign{
		{
			1,
			"test1",
			[]campaign.PositionSetup{{10, 2}, {45, 10}},
			[]int{1, 2, 3, 4, 5},
			3.33,
		},
		{
			2,
			"test2",
			[]campaign.PositionSetup{{20, 2}, {45, 10}},
			[]int{1, 7, 8, 5, 9},
			3.34,
		},
	}
	exampleCampaignsEmpty := []campaign.Campaign{
		{
			1,
			"test1",
			make([]campaign.PositionSetup, 0),
			[]int{1, 2, 3, 4, 5},
			3.33,
		},
		{
			2,
			"test2",
			[]campaign.PositionSetup{{20, 2}, {45, 10}},
			make([]int, 0),
			3.34,
		},
	}

	// Bidder
	tester.Run("Biddder Test", func(test *testing.T) {
		testCases := []bidderCase{
			{
				"Empty campaigns",
				baseSearcher,
				candSearcherMock,
				pubId,
				pos,
				make([]campaign.Campaign, 0),
				// price > 0
				false,
			},
			{
				"Not matching any campaign",
				baseSearcher,
				properSearcher,
				33333333,
				4444444,
				exampleCampaigns,
				// price > 0
				false,
			},
			{
				"Matching from empty conditions",
				baseSearcher,
				properSearcher,
				pubId,
				19,
				exampleCampaignsEmpty,
				// price > 0
				true,
			},
			{
				"Proper matching conditions",
				baseSearcher,
				properSearcher,
				pubId,
				36,
				exampleCampaigns,
				// price > 0
				true,
			},
		}
		for _, tc := range testCases {
			test.Run(tc.name, func(t *testing.T) {
				res := bidder.FindBestBid(tc.pos, tc.pubId, tc.data, tc.candidate, tc.searcher)
				maxPriceSet := res > 0
				assert.True(t, maxPriceSet == tc.expVal, "Got result: %v, Exp: %v", maxPriceSet, tc.expVal)
				if maxPriceSet {
					// MaxPrice set along the tests
					assert.True(t, res == 3.34, "Got result: %v, Exp: %v", res, 3.34)
				}
			})
		}
	})
}
