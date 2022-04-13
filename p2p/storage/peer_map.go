package storage

import (
	"sync"
)

//peerMap is a structure for peer infomation
type peerMap struct {
	l sync.Mutex
	m map[string]*peerInfomation
}

// Store sets the value for a key.
func (n *peerMap) len() int {
	return len(n.m)
}

// Store sets the value for a key.
func (n *peerMap) store(key string, value *peerInfomation) {
	n.l.Lock()
	if 0 == len(n.m) {
		n.m = map[string]*peerInfomation{
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
func (n *peerMap) load(key string) (*peerInfomation, bool) {
	n.l.Lock()
	defer n.l.Unlock()
	v, has := n.m[key]
	return v, has
}

// Delete deletes the value for a key.
func (n *peerMap) delete(key string) {
	n.l.Lock()
	defer n.l.Unlock()

	delete(n.m, key)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (n *peerMap) Range(f func(string, *peerInfomation) bool) {
	n.l.Lock()
	defer n.l.Unlock()

	for key, value := range n.m {
		if !f(key, value) {
			break
		}
	}
}
