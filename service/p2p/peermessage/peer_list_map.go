package peermessage

import (
	"sync"
	"time"
)

//ScoreBoardMap is the structure that stores the ping time of each peer.
type ScoreBoardMap struct {
	l sync.Mutex
	m map[string]time.Duration
}

// Len returns the length of this map.
func (n *ScoreBoardMap) Len() int {
	n.l.Lock()
	defer n.l.Unlock()
	return len(n.m)
}

// Store sets the value for a key.
func (n *ScoreBoardMap) Store(key string, value time.Duration) {
	n.l.Lock()
	if 0 == len(n.m) {
		n.m = map[string]time.Duration{
			key: value,
		}
	} else {
		n.m[key] = value
	}
	n.l.Unlock()
}

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (n *ScoreBoardMap) Load(key string) (time.Duration, bool) {
	n.l.Lock()
	defer n.l.Unlock()
	v, has := n.m[key]
	return v, has
}

// Delete deletes the value for a key.
func (n *ScoreBoardMap) Delete(key string) {
	n.l.Lock()
	defer n.l.Unlock()

	delete(n.m, key)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (n *ScoreBoardMap) Range(f func(string, time.Duration) bool) {
	n.l.Lock()
	defer n.l.Unlock()

	for key, value := range n.m {
		if !f(key, value) {
			break
		}
	}
}
