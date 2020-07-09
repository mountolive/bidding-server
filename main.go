package main

import (
	"fmt"
	"github.com/mountolive/bidding-server/campaign"
	"github.com/mountolive/bidding-server/search"
)

func main() {
	val, err := campaign.RetrieveCampaigns()
	if err != nil {
		fmt.Printf("%s \n", err)
	}
	fmt.Printf("%v \n", val[1].Publishers)

	fmt.Printf("%v \n", search.PublisherAllowed(3875, val[1]))
}
