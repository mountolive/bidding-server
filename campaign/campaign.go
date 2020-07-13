package campaign

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
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

// Interface for holding parsing and retrieval methods
type CampaignRetriever interface {
	// For in-memory load
	getRawCampaigns(redisUrl, hashName string) (map[string]string, error)
	// For "on-flight" retrieval
	getSingleCampaign(ctx context.Context, rdb *redis.Client, hashName, id string) (string, error)
	// Count for "on-fligth" retrieval
	getHashSize(ctx context.Context, rdb *redis.Client, hashName string) int64
	parseCampaign(id, hash string) (*Campaign, error)
}

// Default retriever
type retriever struct{}

// Wrapper function that loads the campaigns from
// default parameters of the ser
func LoadDefaultCampaigns() ([]Campaign, error) {
	return RetrieveCampaigns(retriever{})
}

// Traverse all redis' keys from parent hashName, parse each and return it
// in a channel
func TraverseCampaigns(ctx context.Context, rdb *redis.Client) <-chan Campaign {
	campChan := make(chan Campaign)
	retriever := retriever{}
	go func() {
		defer close(campChan)
		n := retriever.getHashSize(ctx, rdb, HashName)
		var cmpg Campaign
		var err error
		var i int64
		for i = 0; i < n; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				cmpg, err = RetrieveSingleCampaign(ctx, rdb, retriever, fmt.Sprintf("%d", i))
				// Error hiding - Should be fixed. FIXME
				if err != nil {
					// We'll continue for the rest of ids
					// So we just print the error and skip
					fmt.Printf("Error occurred on Traversing for %v, %v \n", i, err)
				}
				select {
				case <-ctx.Done():
					return
				case campChan <- cmpg:
				}
			}
		}
	}()
	return campChan
}

// Retrieves a single campaign by id and parse it
// This one receives the client as it will be reusable
func RetrieveSingleCampaign(ctx context.Context, rdb *redis.Client, cRetr CampaignRetriever, id string) (Campaign, error) {
	raw, err := cRetr.getSingleCampaign(ctx, rdb, HashName, id)
	if err != nil {
		return Campaign{}, err
	}
	campaign, err := cRetr.parseCampaign(id, raw)
	if err != nil {
		return Campaign{}, err
	}
	return *campaign, nil
}

// Function that deals with the retrieval and parsing of campaigns
// Passing the retrieval function so that it can be mocked out
// returns a slice of Campaign
func RetrieveCampaigns(campRetriever CampaignRetriever) ([]Campaign, error) {
	raw, err := campRetriever.getRawCampaigns(RedisUrl, HashName)
	// Bubbling the error up
	if err != nil {
		return nil, err
	}
	// Unmarshaling one by one and appending to slice
	campaigns := make([]Campaign, 0)
	for id, hash := range raw {
		placeholder, err := campRetriever.parseCampaign(id, hash)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, *placeholder)
	}
	return campaigns, nil
}

// Implementations for default CampaignRetriever

// Gets standard map[string]string from redis' query
// of the string-hash related to the campaigns
func (r retriever) getRawCampaigns(redisUrl, hashName string) (map[string]string, error) {
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

// To be used for on-the flight retrieval (as opposed to getRawCamapaigns which retrieves
// the whole key-space
func (r retriever) getSingleCampaign(ctx context.Context, rdb *redis.Client, hashName, id string) (string, error) {
	fmt.Printf("Getting id: %v \n", id)

	raw, err := rdb.HGet(ctx, hashName, id).Result()
	// Error finding hash - Or client closed
	if err != nil {
		fmt.Printf("An error occurred when trying to retrieve cached data for id: %v. Error: %v . Reattempting \n", id, err)
		if err == redis.Nil {
			// We try to reopoen the client (not the best practice)
			rdb, err = NewRedisClient(ctx, RedisUrl)
			if err != nil {
				fmt.Printf("Unable to regenerate client, abborting call. Id: %v. Error: %v \n", id, err)
				return "", err
			}
		} else {
			return "", err
		}
	}
	if raw == "" {
		msg := fmt.Sprintf("The campaigns' for %v hash, id: %v was not found", hashName, id)
		fmt.Println(msg)
		return "", errors.New(msg)
	}
	return raw, nil
}

// Helper function that parses a string into a campaign pointer
func (r retriever) parseCampaign(id, hash string) (*Campaign, error) {
	placeholder := &Campaign{}
	err := json.Unmarshal([]byte(hash), placeholder)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("An error occurred marshaling campaign of id %s: %v \n", id, err))
	}
	return placeholder, nil
}

// Get's the total number of keys for a hash
func (r retriever) getHashSize(ctx context.Context, rdb *redis.Client, hashName string) int64 {
	return rdb.HLen(ctx, hashName).Val()
}
