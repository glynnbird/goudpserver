package main

// Bucket is a "leaky bucket" it has a Capacity (it's maximum size) and a Value (
// it's current size). It is "reset" periodically, which puts the Value equal to the Capacity.
// When the Bucket is "dec"'d the Value is decremented by another number - in this operation,
// there is an opportunity to first set or subsequently set the bucket's capacity too.
type Bucket struct {
	Value    int `json:"value"`
	Capacity int `json:"capacity"`
}

// dec decrements the Bucket's value by "by", or sets it to zero if there isn't enough to
// value left. The return value indicates whether there was enough value left in the bucket
// to decrement. If, say, the Value is 3 and we ask to decrement by 4, the Value will be
// set to 0 and the return value will be true.
func (b *Bucket) dec(by int, capacity int) bool {
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
		b.Value = 0
		return true
	}
}

// reset sets the Value of the bucket to its Capacity
func (b *Bucket) reset() {
	b.Value = b.Capacity
}

// setCapacity changes the Capacity of the bucket.
func (b *Bucket) setCapacity(capacity int) {
	b.Capacity = capacity
}
