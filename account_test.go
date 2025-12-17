package main

import "testing"

func Test_account_new(t *testing.T) {
	accName := "xyz"
	acc := NewAccount(accName)

	// ensure we have one bucket for each classType
	numClassTypes := len(classTypes)
	if len(acc.Buckets) != numClassTypes {
		t.Errorf("Expected buckets to have %v length, got %v", numClassTypes, len(acc.Buckets))
	}

	// check each key
	for _, v := range classTypes {
		_, ok := acc.Buckets[v]
		if !ok {
			t.Errorf("Expected buckets to have a key %v, but it is missing", v)
		}
	}

	// check account name
	if acc.Name != accName {
		t.Errorf("Expected account to have a name %v, but got %v", accName, acc.Name)
	}
}

func Test_account_reset(t *testing.T) {
	accName := "zyx"
	acc := NewAccount(accName)
	acc.Buckets["l"].set(50, 100)
	acc.Buckets["w"].set(25, 50)
	acc.Buckets["q"].set(2, 5)
	acc.reset()
	if acc.Buckets["l"].Value() != 100 {
		t.Errorf("Expected l bucket to have a value %v, but got %v", 100, acc.Buckets["l"].Value())
	}
	if acc.Buckets["w"].Value() != 50 {
		t.Errorf("Expected w bucket to have a value %v, but got %v", 50, acc.Buckets["w"].Value())
	}
	if acc.Buckets["q"].Value() != 5 {
		t.Errorf("Expected q bucket to have a value %v, but got %v", 5, acc.Buckets["w"].Value())
	}
}
