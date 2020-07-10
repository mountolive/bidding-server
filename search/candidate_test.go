package search

import (
	"github.com/mountolive/bidding-server/campaign"
	"github.com/stretchr/testify/assert"
	"testing"
)

type candidateCase struct {
	name    string
	compare int
	data    interface{}
	expVal  bool
	expErr  bool
}

// Just base tests
// searcher can be mocked by creating a stub searcher
func TestCandidate(tester *testing.T) {
	candSrchr := candidateSearcher{}
	id := 1
	exampleIntTrue := []int{6, 4, 3, 4, 52, 1, 4}
	examplePosTrue := []campaign.PositionSetup{{18, 3}, {43, 3}, {7, 6}}
	exampleIntFalse := []int{6, 4, 3, 4, 52, 8, 4}
	examplePosFalse := []campaign.PositionSetup{{18, 3}, {43, 3}, {8, 6}}
	// Candidate
	tester.Run("Candidate Test", func(test *testing.T) {
		testCases := []candidateCase{
			{
				"Wrong data type",
				id,
				"not valid",
				false,
				// err != nil
				true,
			},
			{
				"Empty int array",
				id,
				make([]int, 0),
				true,
				// err != nil
				false,
			},
			{
				"Empty PositionSetup",
				id,
				make([]campaign.PositionSetup, 0),
				true,
				// err != nil
				false,
			},
			{
				"Not found element int",
				id,
				make([]int, 1),
				false,
				// err != nil
				false,
			},
			{
				"Not found element positionSetup",
				id,
				make([]campaign.PositionSetup, 1),
				false,
				false,
			},
			// The following test cases should be really part of
			// a test suite for the private methods
			{
				"Found int proper example",
				id,
				exampleIntTrue,
				true,
				false,
			},
			{
				"Found position proper example",
				id,
				examplePosTrue,
				true,
				false,
			},
			{
				"Not found int proper example",
				id,
				exampleIntFalse,
				false,
				false,
			},
			{
				"Not found position proper example",
				id,
				examplePosFalse,
				false,
				false,
			},
		}
		for _, tc := range testCases {
			test.Run(tc.name, func(t *testing.T) {
				res, err := candSrchr.Candidate(tc.compare, tc.data)
				noErr := err != nil
				assert.True(t, tc.expVal == res, "Got result: %v, Exp: %v", res, tc.expVal)
				assert.True(t, tc.expErr == noErr, "Got err: %v, Exp: %v --- %v", noErr, tc.expErr, err)
			})
		}
	})
}
