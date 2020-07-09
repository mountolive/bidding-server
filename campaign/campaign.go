package campaign

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Struct that represents a single campaign
// Publishers will get parsed as a float64 array
// id is also
type Campaign struct {
	id         int             `json:"id"`
	name       string          `json:"name"`
	Positions  []PositionSetup `json:"positions"`
	Publishers []int           `json:"publishers"`
	Cpm        float64         `json:"cpm"`
}

// Struct that represents a positionSetup element of a campaign
// Setting float64 is admittedly an overkill for memory
// but makes it easier for code reuse
type PositionSetup struct {
	Position int `json:"position"`
	Distance int `json:"distance"`
}

// Function that deals with the retrieval and parsing of campaigns
// returns a slice of Campaign
func RetrieveCampaigns() ([]Campaign, error) {
	raw, err := getRawCampaigns(RedisUrl, HashName)
	// Bubbling the error up
	if err != nil {
		return nil, err
	}
	// Unmarshaling one by one and appending to slice
	campaigns := make([]Campaign, 0)
	placeholder := &Campaign{}
	for id, hash := range raw {
		err := json.Unmarshal([]byte(hash), placeholder)
		campaigns = append(campaigns, *placeholder)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("An error occurred marshaling campaign of id %s: %v \n", id, err))
		}
	}
	return campaigns, nil
}

// Gets standard map[string]string from redis' query
// of the string-hash related to the campaigns
func getRawCampaigns(redisUrl, hashName string) (map[string]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rdb, err := NewRedisClient(ctx, redisUrl)
	// We will close connection because we won't be needing it again
	// (Although, in any case we're cancelling the context before)
	defer rdb.Close()
	// It was an error when connecting to redis
	if err != nil {
		fmt.Println("An error occurred when connecting to redis")
		return nil, err
	}
	raw, err := rdb.HGetAll(ctx, hashName).Result()
	// Error finding hash
	if err != nil {
		fmt.Println("An error occurred when trying to retrieve cached data")
		return nil, err
	}
	// Non-existing hash
	if len(raw) == 0 {
		msg := "The campaigns' hash was not found"
		fmt.Println(msg)
		return nil, errors.New(msg)
	}
	return raw, nil
}
