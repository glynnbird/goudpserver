package main

import "testing"

func Test_account_map_new(t *testing.T) {
	am := NewAccountMap()
	if am.accounts == nil {
		t.Error("Expected accounts map to have been initialised but got nil")
	}
}

func Test_account_map_LoadOrStore_add_missing(t *testing.T) {
	am := NewAccountMap()
	am.LoadOrStore("bob")
	if len(am.accounts) != 1 {
		t.Errorf("Expected accounts map to have length of 1, got %v", len(am.accounts))
	}
}

func Test_account_map_LoadOrStore_dedupe(t *testing.T) {
	am := NewAccountMap()
	am.LoadOrStore("bob")
	am.accounts["bob"].Buckets["l"].dec(1, 100)
	am.accounts["bob"].Buckets["w"].dec(1, 50)
	am.accounts["bob"].Buckets["q"].dec(1, 5)
	am.LoadOrStore("bob")
	if len(am.accounts) != 1 {
		t.Errorf("Expected accounts map to have length of 1, got %v", len(am.accounts))
	}
	// test we got the original account not a new one after fetching "bob" twice
	if am.accounts["bob"].Buckets["l"].Capacity() != 100 {
		t.Errorf("Expected account's bucket capacity to be 100, got %v", am.accounts["bob"].Buckets["l"].Capacity())
	}
	if am.accounts["bob"].Buckets["w"].Capacity() != 50 {
		t.Errorf("Expected account's bucket capacity to be 50, got %v", am.accounts["bob"].Buckets["w"].Capacity())
	}
	if am.accounts["bob"].Buckets["q"].Capacity() != 5 {
		t.Errorf("Expected account's bucket capacity to be 5, got %v", am.accounts["bob"].Buckets["q"].Capacity())
	}
	if am.accounts["bob"].Buckets["l"].Value() != 99 {
		t.Errorf("Expected account's bucket value to be 99, got %v", am.accounts["bob"].Buckets["l"].Value())
	}
	if am.accounts["bob"].Buckets["w"].Value() != 49 {
		t.Errorf("Expected account's bucket value to be 49, got %v", am.accounts["bob"].Buckets["w"].Value())
	}
	if am.accounts["bob"].Buckets["q"].Value() != 4 {
		t.Errorf("Expected account's bucket value to be 4, got %v", am.accounts["bob"].Buckets["q"].Value())
	}
}

func Test_account_map_LoadOrStore_multi(t *testing.T) {
	am := NewAccountMap()
	am.LoadOrStore("bob")
	am.LoadOrStore("rita")
	am.LoadOrStore("rita")
	am.LoadOrStore("bob")
	am.LoadOrStore("rita")
	am.LoadOrStore("sue")
	am.LoadOrStore("bob")
	am.LoadOrStore("bob")
	if len(am.accounts) != 3 {
		t.Errorf("Expected accounts map to have length of 3, got %v", len(am.accounts))
	}
}

func Test_account_map_reset(t *testing.T) {
	am := NewAccountMap()
	am.LoadOrStore("bob")
	am.accounts["bob"].Buckets["l"].dec(1, 100)
	am.accounts["bob"].Buckets["w"].dec(1, 50)
	am.accounts["bob"].Buckets["q"].dec(1, 5)
	am.LoadOrStore("rita")
	am.accounts["rita"].Buckets["l"].dec(1, 100)
	am.accounts["rita"].Buckets["w"].dec(1, 50)
	am.accounts["rita"].Buckets["q"].dec(1, 5)
	am.LoadOrStore("sue")
	am.accounts["sue"].Buckets["l"].dec(1, 100)
	am.accounts["sue"].Buckets["w"].dec(1, 50)
	am.accounts["sue"].Buckets["q"].dec(1, 5)
	am.Reset()
	for accName, acc := range am.accounts {
		if acc.Buckets["l"].Capacity() != 100 {
			t.Errorf("Expected account %v to have %v capacity of %v, got %v", accName, "l", 100, acc.Buckets["l"].Capacity())
		}
		if acc.Buckets["w"].Capacity() != 50 {
			t.Errorf("Expected account %v to have %v capacity of %v, got %v", accName, "w", 50, acc.Buckets["w"].Capacity())
		}
		if acc.Buckets["q"].Capacity() != 5 {
			t.Errorf("Expected account %v to have %v capacity of %v, got %v", accName, "q", 5, acc.Buckets["q"].Capacity())
		}
		if acc.Buckets["l"].Capacity() != acc.Buckets["l"].Value() {
			t.Errorf("Expected account %v to have %v value same as its capacity, got %v", accName, "l", acc.Buckets["l"].Value())
		}
		if acc.Buckets["w"].Capacity() != acc.Buckets["w"].Value() {
			t.Errorf("Expected account %v to have %v value same as its capacity, got %v", accName, "w", acc.Buckets["w"].Value())
		}
		if acc.Buckets["q"].Capacity() != acc.Buckets["q"].Value() {
			t.Errorf("Expected account %v to have %v value same as its capacity, got %v", accName, "q", acc.Buckets["q"].Value())
		}
	}
}
