package search

import (
	"context"
	"github.com/mountolive/bidding-server/campaign"
	"runtime"
	"sync"
)

// Checks whether the publisher can bid on this campaign
func PublisherAllowed(pubId int, cmpg campaign.Campaign) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Open to any publisher
	pubs := cmpg.Publishers
	n := len(pubs)
	if n == 0 {
		return true
	}
	// Possible "workers"
	cpus := runtime.NumCPU()
	// Work splitter
	k := n / cpus
	// Offset pointer
	bottom := 0
	top := k
	finders := make([]<-chan bool, cpus)
	for i := 0; i < cpus; i++ {
		finders[i] = linearSearch(ctx, pubId, pubs[bottom:top])
		bottom = top
		top += k
		// Reached the end of the array
		if top > n {
			top = n
		}
	}
	// Executing actual search
	for val := range fanIn(ctx, finders...) {
		if val {
			return true
		}
	}
	// We didn't find the publisher
	return false
}

// Returns the comparison of each element in the array
// to the publisher's id, in the form of a chan
func linearSearch(ctx context.Context, pubId int, ids []int) <-chan bool {
	found := make(chan bool)
	go func() {
		defer close(found)
		for _, i := range ids {
			select {
			case <-ctx.Done():
				return
			case found <- (i == pubId):
			}
		}
	}()
	return found
}

// Converts a varags of chans into a single one
func fanIn(ctx context.Context, chans ...<-chan bool) <-chan bool {
	multiplexer := make(chan bool)
	var wg sync.WaitGroup
	// Size of the multiplexer
	n := len(chans)
	multiplex := func(c <-chan bool) {
		defer wg.Done()
		for val := range c {
			select {
			case <-ctx.Done():
				return
			case multiplexer <- val:
			}
		}
	}
	// All channels (in this case, numOfCpus)
	wg.Add(n)
	for _, c := range chans {
		go multiplex(c)
	}

	// Waiting and closing
	go func() {
		wg.Wait()
		close(multiplexer)
	}()
	return multiplexer
}
