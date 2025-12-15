package main

// Account is a data structure that stores everything we need to know
// about a user account: its name and three leaky buckets for lookups,
// writes and queries, modelling a rate-limited API that has three
// separate quotas for reads/writes/queries per second
type Account struct {
	Name    string             `json:"name"`
	Buckets map[string]*Bucket `json:"buckets"`
}

// NewAccount creates a new account given the new account's name.
func NewAccount(name string) *Account {
	lookups := Bucket{}
	writes := Bucket{}
	queries := Bucket{}
	buckets := map[string]*Bucket{
		"l": &lookups,
		"w": &writes,
		"q": &queries,
	}
	acc := Account{
		Name:    name,
		Buckets: buckets,
	}
	return &acc
}

// reset sets each leaky bucket back to its full capacity
func (acc *Account) reset() {
	for _, b := range acc.Buckets {
		b.reset()
	}
}
