package main

import (
	"sync"
)

// Bucket is a "leaky bucket" it has a Capacity (it's maximum size) and a Value (
// it's current size). It is "reset" periodically, which puts the Value equal to the Capacity.
// When the Bucket is "dec"'d the Value is decremented by another number - in this operation,
// there is an opportunity to first set or subsequently set the bucket's capacity too.
type Bucket struct {
	Value    int `json:"value"`
	Capacity int `json:"capacity"`
	mu       sync.Mutex
}

// dec decrements the Bucket's value by "by", or returns false if there isn't enough value left.
// The return value indicates whether there was enough value left in the bucket
// to decrement. Some scenarios:
// - "Value" is 10 and "by" is 1. Value is set to 9 and return is true
// - "Value" is 1 and "by" is 1. Value is set to 0 and return is true
// - "Value" is 0 and "by" is 1. Value stays set to 0 and return is false
// - "Value" is 1 and "by" is 2. Value stays set to 1 and return is false
// The bucket size is passed in and set every time.
func (b *Bucket) dec(by int, capacity int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if by <= 0 || capacity <= 0 {
		return false
	}
	if b.Capacity == 0 {
		b.Value = capacity
	}
	b.Capacity = capacity
	if b.Value == 0 {
		return false
	}
	if b.Value >= by {
		b.Value -= by
		return true
	} else {
		return false
	}
}

// reset sets the Value of the bucket to its Capacity
func (b *Bucket) reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Value = b.Capacity
}
