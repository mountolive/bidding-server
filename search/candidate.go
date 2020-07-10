package search

import (
	"context"
	"errors"
	"github.com/mountolive/bidding-server/campaign"
	"math"
	"runtime"
	"sync"
)

// Interface that defines our search methods
type Search interface {
	searchPublishers(ctx context.Context, pubId int, data []int) <-chan bool
	searchPositions(ctx context.Context, position int, data []campaign.PositionSetup) <-chan bool
}

// Implementation holder
type searcher struct{}

type SearchCandidate interface {
	Candidate(ctx context.Context, i int, data interface{}, srchr Search) (bool, error)
}

type candidateSearcher struct{}

// Checks whether the publisher can bid on this campaign
// Depending on the type of data, it would search for positions or
// permitted publishers. If the data is not supported,
// err will be not nil
func (s candidateSearcher) Candidate(ctx context.Context, pubValue int, data interface{}, searcher Search) (bool, error) {
	//ctx, cancel := context.WithCancel(ctx)
	//defer cancel()
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
	// Possible "workers"
	cpus := runtime.NumCPU()
	// Work splitter
	var k int
	if n > cpus {
		k = int(math.Ceil(float64(n) / float64(cpus)))
	} else {
		// One comparison per core
		k = 1
	}
	// Offset pointer
	bottom := 0
	top := k
	finders := make([]<-chan bool, cpus)
	for i := 0; i < cpus; i++ {
		// We're trying to find publishers
		if publishers != nil {
			finders[i] = searcher.searchPublishers(ctx, pubValue, publishers[bottom:top])
		} else {
			// We're trying to find positions
			finders[i] = searcher.searchPositions(ctx, pubValue, positions[bottom:top])
		}
		if top == n {
			// No need for creating more workers
			break
		}
		bottom = top
		top += k
		// Reached the end of the array
		if top > n {
			top = n
		}
	}
	// Executing actual search
	// We're accepting up to n possible results
	for val := range take(ctx, fanIn(ctx, finders...), n) {
		if val {
			return true, nil
		}
	}
	// We didn't find the publisher or position, so, no candidate
	return false, nil
}

// The following 2 methods might be candidates to be handled by one
// generic method

// Returns the comparison of each element in the array
// to the publisher's id, in the form of a chan (linear execution)
func (s searcher) searchPublishers(ctx context.Context, pubId int, data []int) <-chan bool {
	found := make(chan bool, 1)
	go func() {
		defer close(found)
		for _, i := range data {
			select {
			case <-ctx.Done():
				return
			case found <- (i == pubId):
			}
		}
	}()
	return found
}

// Returns the comparison of each element in the array
// to the publisher's position, in the form of a boolean chan (linear)
func (s searcher) searchPositions(ctx context.Context, position int, data []campaign.PositionSetup) <-chan bool {
	found := make(chan bool, 1)
	cmp := func(pos int, arrPos campaign.PositionSetup) bool {
		// shortening the names
		dist := arrPos.Distance
		currPos := arrPos.Position
		return currPos-dist <= pos && pos <= currPos+dist
	}
	go func() {
		defer close(found)
		for _, i := range data {
			select {
			case <-ctx.Done():
				return
			case found <- cmp(position, i):
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

// Consumes up to num iterations to stream
func take(ctx context.Context, stream <-chan bool, num int) <-chan bool {
	takeStream := make(chan bool)
	go func() {
		defer close(takeStream)
		for i := 0; i < num; i += 1 {
			select {
			case <-ctx.Done():
				return
			case takeStream <- <-stream:
			}
		}
	}()
	return takeStream
}
