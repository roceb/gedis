package server

import (
	"sync"
)

type gedis struct {
	items  map[string]string // used to store all of the Key Value pairs
	mu     sync.RWMutex      // used to lock gedis to avoid race condition
	counts map[string]int    // used to keep track of count
	trans  map[int][]string  // tranaxtions are stored here
}

// creates an empty database
func newDB() gedis {
	items := map[string]string{}
	counters := map[string]int{}
	trans := map[int][]string{}
	return gedis{items: items, counts: counters, trans: trans}
}

// command adds [key] [value] to gedis.items and add [value] to gedis.count
func (g *gedis) set(key, value string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if val, ok := g.items[key]; ok {
		g.counts[val]--
	}
	g.items[key] = value
	g.counts[value]++
}

// needs a [key] and  returns [value] and err. Fetched from gedis
func (g *gedis) get(key string) (string, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	value, exist := g.items[key]
	return value, exist
}

// deletes a [key] and lowers the count in gedis.counts
func (g *gedis) delete(key string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	v := g.items[key]
	delete(g.items, key)
	g.counts[v]--

}

// returns the amount of key in gedis
func (g *gedis) count(key string) int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.counts[key]
}

// informs user that tranaction is starting
func (g *gedis) begin() string {
	return "Queue started"
}

// deletes last transaction
func (g *gedis) rollback(id int) (rollback string) {
	rollback = "DELETING COMMIT"
	if len(g.trans) <= 0 {
		rollback = "TRANSACTION NOT FOUND"
	}
	delete(g.trans, id)

	return rollback
}

// did not have enough time to move logic in here for commit command.
// func (g *gedis) commit() {

// }
