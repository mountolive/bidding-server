package campaign

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type getRawCampaignCase struct {
	name   string
	addr   string
	hash   string
	expVal bool
	expErr bool
}

type retrieveCampaignsCase struct {
	name   string
	getRaw func(url, hash string) (map[string]string, error)
	expVal bool
	expErr bool
}

// By no means a complete suite. Just base cases
func TestCampaign(tester *testing.T) {
	// RetrieveCampaigns
	tester.Run("RetrieveCampaigns Test", func(test *testing.T) {
		badParsing := func(u, h string) (map[string]string, error) {
			res := make(map[string]string)
			res["test"] = "this can't be parsed"
			return res, nil
		}
		connErr := func(u, h string) (map[string]string, error) {
			return nil, errors.New("Connection error")
		}
		testCases := []retrieveCampaignsCase{
			{
				"Parsing error",
				badParsing,
				false,
				true,
			},
			{
				"Connectivity error",
				connErr,
				false,
				true,
			},
			{
				"Proper retrieval of data",
				getRawCampaigns,
				true,
				false,
			},
		}
		for _, tc := range testCases {
			test.Run(tc.name, func(t *testing.T) {
				res, err := RetrieveCampaigns(tc.getRaw)
				noRes := res != nil
				noErr := err != nil
				assert.True(t, tc.expVal == noRes, "Got result: %s, Exp: %s", noRes, tc.expVal)
				assert.True(t, tc.expErr == noErr, "Got err: %s, Exp: %s", noErr, tc.expErr)
				// Minor assertions regarding data
				if noRes && tc.expVal {
					assert.True(t, len(res) > 0, "Wrongly parsed data returned empty array")
				}
			})
		}
	})

	// GetRawCampaigns
	tester.Run("getRawCampaigns Test", func(test *testing.T) {
		// Definition
		testCases := []getRawCampaignCase{
			{
				"Connectivity error",
				"localhost:678",
				HashName,
				false,
				true,
			},
			{
				"Not existent or empty hash",
				RedisUrl,
				"non-existent",
				false,
				true,
			},
			{
				"Existing data hash",
				RedisUrl,
				HashName,
				true,
				false,
			},
		}
		// Execution
		for _, tc := range testCases {
			test.Run(tc.name, func(t *testing.T) {
				res, err := getRawCampaigns(tc.addr, tc.hash)
				noRes := res != nil
				noErr := err != nil
				assert.True(t, tc.expVal == noRes, "Got result: %s, Exp: %s", noRes, tc.expVal)
				assert.True(t, tc.expErr == noErr, "Got err: %s, Exp: %s", noErr, tc.expErr)
			})
		}
	})

}
