package main

import "testing"

func Test_bucket_dec_with_enough_value(t *testing.T) {
	bucket := &Bucket{}
	bucket.set(10, 10)
	permitted := bucket.dec(1, 10)
	if !permitted {
		t.Errorf("Expected permitted to be true, got false")
	}
	if bucket.Value() != 9 {
		t.Errorf("Expected bucket Value to be 9, got %d", bucket.Value())
	}
}

func Test_bucket_dec_without_enough_value(t *testing.T) {
	bucket := &Bucket{}
	bucket.set(1, 10)
	permitted := bucket.dec(2, 10)
	if permitted {
		t.Errorf("Expected permitted to be false, got true")
	}
	if bucket.Value() != 1 {
		t.Errorf("Expected bucket Value to remain 1, got %d", bucket.Value())
	}
}

func Test_bucket_dec_with_zero_value(t *testing.T) {
	bucket := &Bucket{}
	bucket.set(0, 10)
	permitted := bucket.dec(1, 10)
	if permitted {
		t.Errorf("Expected permitted to be false, got true")
	}
	if bucket.Value() != 0 {
		t.Errorf("Expected bucket Value to remain 0, got %d", bucket.Value())
	}
}

func Test_bucket_dec_with_capacity_change(t *testing.T) {
	bucket := &Bucket{}
	bucket.set(10, 10)
	permitted := bucket.dec(2, 100)
	if permitted == false {
		t.Errorf("Expected permitted to be true, got false")
	}
	if bucket.Value() != 8 {
		t.Errorf("Expected bucket Value to be 8, got %d", bucket.Value())
	}
	if bucket.Capacity() != 100 {
		t.Errorf("Expected bucket Capacity to be 100, got %d", bucket.Capacity())
	}
}
func Test_bucket_set_success(t *testing.T) {
	bucket := &Bucket{}
	bucket.set(44, 55)
	if bucket.Value() != 44 {
		t.Errorf("Expected bucket Value to be 44, got %d", bucket.Value())
	}
	if bucket.Capacity() != 55 {
		t.Errorf("Expected bucket Value to be 55, got %d", bucket.Capacity())
	}
}

func Test_bucket_set_negative_value(t *testing.T) {
	bucket := &Bucket{}
	err := bucket.set(-1, 55)
	if err == nil {
		t.Error("Expected error for setting negative value, got nil")
	}
}

func Test_bucket_set_negative_capacity(t *testing.T) {
	bucket := &Bucket{}
	err := bucket.set(22, -22)
	if err == nil {
		t.Error("Expected error for setting negative capacity, got nil")
	}
}

func Test_bucket_set_value_more_than_capacity(t *testing.T) {
	bucket := &Bucket{}
	err := bucket.set(101, 100)
	if err == nil {
		t.Error("Expected error for setting value more than capacity, got nil")
	}
}

func Test_bucket_reset(t *testing.T) {
	bucket := &Bucket{}
	bucket.set(5, 10)
	bucket.reset()
	if bucket.Value() != 10 {
		t.Errorf("Expected bucket Value to be 10, got %d", bucket.Value())
	}
}
