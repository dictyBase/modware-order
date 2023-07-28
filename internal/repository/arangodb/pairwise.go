package arangodb

import (
	"errors"

	"github.com/dictyBase/modware-order/internal/model"
)

// PairWiseIterator is the container for iterator.
type PairWiseIterator struct {
	slice []*model.OrderDoc
	// keeps track of the first index
	firstIdx int
	// keeps track of the next index in the pair
	secondIdx int
	// last index of the slice
	lastIdx int
	// toogle the state for fetching the first pair
	firstPair bool
}

// NewPairWiseIterator is the constructor, returns error in case of empty or
// slice with single element.
func NewPairWiseIterator(mrd []*model.OrderDoc) (*PairWiseIterator, error) {
	if len(mrd) <= 1 {
		return &PairWiseIterator{}, errors.New("not enough element to fetch pairs")
	}

	return &PairWiseIterator{
		slice:     mrd,
		firstIdx:  0,
		secondIdx: 1,
		lastIdx:   len(mrd) - 1,
		firstPair: true,
	}, nil
}

// NextPair moves the iteration to the next pair. If NextPair() returns true
// the pair could be retrieved by Pair() method. If it is called for the first
// time it points to the first pair.
func (p *PairWiseIterator) NextPair() bool {
	if p.firstPair {
		p.firstPair = false

		return true
	}
	if p.secondIdx == p.lastIdx {
		return false
	}
	p.firstIdx++
	p.secondIdx++

	return true
}

// Pair retrieves the pair of elements from the slice.
func (p *PairWiseIterator) Pair() (*model.OrderDoc, *model.OrderDoc) {
	return p.slice[p.firstIdx], p.slice[p.secondIdx]
}
