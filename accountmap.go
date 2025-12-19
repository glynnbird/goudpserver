package main

import (
	"sync"
)

// AccountMap models a map of Account structs and a Mutex to ensure thread-safety
type AccountMap struct {
	accounts map[string]Account
	mu       sync.RWMutex
}

// NewAccountMap creates a new AccountMap with an empty accounts map.
func NewAccountMap() *AccountMap {
	accounts := make(map[string]Account)
	am := AccountMap{
		accounts: accounts,
	}
	return &am
}

// LoadOrStore fetches an Account from our map of accounts given its accountName. If
// it doesn't exist, a new Account is created and added to the map. All the necessary
// read and write locking is performed for thread-safety.
func (am *AccountMap) LoadOrStore(accountName string) (Account, bool) {
	newAccountCreated := false
	am.mu.RLock()
	acc, ok := am.accounts[accountName]
	am.mu.RUnlock()
	// if the key isn't in our map, create a new Account
	if !ok {
		am.mu.Lock()
		// this is to solve a race between two goroutines trying to create the same map key
		if otherAcc, exists := am.accounts[accountName]; !exists {
			acc = NewAccount(accountName)
			am.accounts[accountName] = acc
			newAccountCreated = true
		} else {
			// the other goroutine beat us to it
			acc = otherAcc
		}
		am.mu.Unlock()
	}
	return acc, newAccountCreated
}

// Reset iterates through our map of accounts, calling "reset" on each account,
// which sets each bucket in each account back to their capacity.
func (am *AccountMap) Reset() {
	am.mu.RLock()
	for _, acc := range am.accounts {
		acc.reset()
	}
	am.mu.RUnlock()
}
