package main

import "sync"

// AccountMap is a wrapper around sync.Map that only accepts strings as keys and Accounts
// as values, for type safety. sync.Map is used to avoid race conditions.
type AccountMap struct {
	m sync.Map
}

// Store wraps arount sync.Map's Store but only accepts string/Account pairs
func (am *AccountMap) Store(key string, value Account) {
	am.m.Store(key, value)
}

// Load wraps arount sync.Map's Load but only accepts string parameters and
// returns an Account
func (am *AccountMap) Load(key string) (Account, bool) {
	v, ok := am.m.Load(key)
	if !ok {
		return *NewAccount(key), false
	}
	s, ok := v.(Account)
	return s, ok
}

// Delete wraps around sync.Map's Delete but only acceptes string keys
func (am *AccountMap) Delete(key string) {
	am.m.Delete(key)
}

// Range wraps around sync.Map's Range
func (am *AccountMap) Range(f func(key string, value Account) bool) {
	am.m.Range(func(k, v any) bool {
		return f(k.(string), v.(Account))
	})
}
