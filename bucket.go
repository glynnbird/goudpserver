package main

import (
	"encoding/json"
	"errors"
	"sync"
)

// Bucket is a "leaky bucket" it has a capacity (it's maximum size) and a calue (
// it's current size). It is "reset" periodically, which puts the value equal to the capacity.
// When the Bucket is "dec"'d the Value is decremented by another number - in this operation,
// there is an opportunity to first set or subsequently set the bucket's capacity too.
type Bucket struct {
	value    int
	capacity int
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
	if b.capacity == 0 {
		b.value = capacity
	}
	b.capacity = capacity

	// if there is sufficient Value left in the bucket
	if b.value >= by {
		// remove Value from the bucket and indicated success
		b.value -= by
		return true
	}
	// otherwise fail
	return false
}

// reset sets the Value of the bucket to its Capacity
func (b *Bucket) reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.value = b.capacity
}

// set sets the value and capacity of the bucket
func (b *Bucket) set(value int, capacity int) error {
	if value < 0 || capacity < 0 {
		return errors.New("cannot accept negative value or capacity")
	}
	if value > capacity {
		return errors.New("value cannot exceed capacity")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.value = value
	b.capacity = capacity
	return nil
}

// Value returns the unexported value attribute
func (b *Bucket) Value() int {
	return b.value
}

// Capacity returns the unexported capacity attribute
func (b *Bucket) Capacity() int {
	return b.capacity
}

// MarshalJSON returns a JSON representation of the bucket's capacity and value
func (b *Bucket) MarshalJSON() ([]byte, error) {
	type Alias struct {
		Value    int `json:"value"`
		Capacity int `json:"capacity"`
	}

	return json.Marshal(Alias{
		Value:    b.value,
		Capacity: b.capacity,
	})
}
