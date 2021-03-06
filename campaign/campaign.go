package campaign

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Struct that represents a single campaign
// Publishers will get parsed as a float64 array
// id is also
type Campaign struct {
	Id         int             `json:"id"`
	Name       string          `json:"name"`
	Positions  []PositionSetup `json:"positions"`
	Publishers []int           `json:"publishers"`
	Cpm        float64         `json:"cpm"`
}

// Struct that represents a positionSetup element of a campaign
type PositionSetup struct {
	Position int `json:"position"`
	Distance int `json:"distance"`
}

// Routine that periodically will update related campaign's data
// It will keep prompting the same campaigns and it would only update them
// after one minute
func GenerateRedisData(ctx context.Context) <-chan []Campaign {
	// Regular chans has no "buffer" size. Meaning it blocks for reading
	// and for writing...
	campaignsChan := make(chan []Campaign)
	go func() {
		defer close(campaignsChan)
		now := time.Now()
		wait := false
		var cmpgs []Campaign
		var err error
		for {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Println("Retrieving redis data")
				// If not wait, reload campaigns
				if !wait {
					fmt.Println("Updating redis data...")
					cmpgs, err = LoadDefaultCampaigns()
					if err != nil {
						fmt.Printf("An error occurred when updating campaigns")
						return
					}
					// Updating the clock
					now = time.Now()
				}
				select {
				case <-ctx.Done():
					return
				// First time we write, we'll trigger future waits
				case campaignsChan <- cmpgs:
					wait = time.Since(now).Minutes() < 1
				}
			}
		}
	}()
	return campaignsChan
}

// Wrapper function that loads the campaigns from
// default parameters of the ser
func LoadDefaultCampaigns() ([]Campaign, error) {
	return RetrieveCampaigns(getRawCampaigns)
}

// Function that deals with the retrieval and parsing of campaigns
// Passing the retrieval function so that it can be mocked out
// returns a slice of Campaign
func RetrieveCampaigns(getRaw func(url, hash string) (map[string]string, error)) ([]Campaign, error) {
	raw, err := getRaw(RedisUrl, HashName)
	// Bubbling the error up
	if err != nil {
		return nil, err
	}
	// Unmarshaling one by one and appending to slice
	campaigns := make([]Campaign, 0)
	for id, hash := range raw {
		placeholder := &Campaign{}
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
	// It was an error when connecting to redis
	if err != nil {
		fmt.Println("An error occurred when connecting to redis")
		return nil, err
	}
	// We will close connection because we won't be needing it again
	// (Although, in any case we're cancelling the context before)
	defer rdb.Close()
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
