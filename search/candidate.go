package search

import (
	"errors"
	"github.com/mountolive/bidding-server/campaign"
)

type SearchCandidate interface {
	Candidate(i int, data interface{}) (bool, error)
}

type candidateSearcher struct{}

// Checks whether the publisher can bid on this campaign
// Depending on the type of data, it would search for positions or
// permitted publishers. If the data is not supported,
// err will be not nil
func (s candidateSearcher) Candidate(pubValue int, data interface{}) (bool, error) {
	// Holder in case data is about publishers
	var publishers []int
	// Holder in case data is about position
	var positions []campaign.PositionSetup
	// We assert the type for each case otherwise we error out
	switch data.(type) {
	case []int:
		publishers = data.([]int)
	case []campaign.PositionSetup:
		positions = data.([]campaign.PositionSetup)
	default:
		return false, errors.New("Passed type for data not supported")
	}
	var n int
	if publishers != nil {
		n = len(publishers)
	} else {
		n = len(positions)
	}
	// Open to any publisher or position
	if n == 0 {
		return true, nil
	}
	// We start the search. Naive linear
	if publishers != nil {
		for _, pub := range publishers {
			if pub == pubValue {
				return true, nil
			}
		}
	} else {
		for _, pos := range positions {
			if positionComparator(pubValue, pos) {
				return true, nil
			}
		}
	}
	// We didn't find the publisher or position, so, no candidate
	return false, nil
}

// Simple comparator to be used for positions matching
func positionComparator(pos int, arrPos campaign.PositionSetup) bool {
	// shortening the names
	dist := arrPos.Distance
	currPos := arrPos.Position
	return currPos-dist <= pos && pos <= currPos+dist
}
