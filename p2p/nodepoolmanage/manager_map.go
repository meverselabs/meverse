package nodepoolmanage

import (
	"bytes"
	"time"

	// "github.com/meverselabs/meverse/p2p/storage"
	"os"
	"path/filepath"
	"sync"

	"github.com/meverselabs/meverse/core/keydb"
	"github.com/meverselabs/meverse/p2p/peermessage"
	"github.com/pkg/errors"
)

//NodeStore is the structure of the connection information.
type nodeStore struct {
	l  sync.Mutex
	db *keydb.DB
	a  []peermessage.ConnectInfo
	m  map[string]*peermessage.ConnectInfo
}

//NewNodeStore is creator of NodeStore
func newNodeStore(dbpath string) (*nodeStore, error) {
	db, err := openNodesDB(dbpath)
	if err != nil {
		return nil, err
	}
	n := &nodeStore{
		db: db,
	}

	go func() {
		for {
			n.shrink()
			time.Sleep(time.Minute)
		}
	}()

	mp := map[string]string{}
	if err := db.Update(func(txn *keydb.Tx) error {
		txn.Ascend(func(key []byte, value interface{}) bool {
			mp[string(key)] = string(value.([]byte))
			return true
		})
		return nil
	}); err != nil {
		return nil, err
	}
	for _, value := range mp {
		bf := bytes.NewBuffer([]byte(value))
		var ci peermessage.ConnectInfo
		ci.ReadFrom(bf)
		ci.PingScoreBoard = &peermessage.ScoreBoardMap{}
		n.LoadOrStore(ci.Address, ci)
	}
	return n, nil
}

func openNodesDB(dbPath string) (*keydb.DB, error) {
	os.MkdirAll(filepath.Dir(dbPath), os.ModePerm)

	db, err := keydb.Open(dbPath, func(key []byte, data []byte) (interface{}, error) {
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (pm *nodeStore) shrink() {
	defer func() {
		recover()
	}()
	pm.db.Shrink()
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (n *nodeStore) LoadOrStore(key string, value peermessage.ConnectInfo) peermessage.ConnectInfo {
	n.l.Lock()
	defer n.l.Unlock()

	if 0 == len(n.m) {
		n.unsafeStore(key, value)
	} else {
		v, has := n.m[key]
		if has {
			return *v
		}
		n.unsafeStore(key, value)
	}
	return value
}

// Store sets the value for a key.
func (n *nodeStore) Store(key string, value peermessage.ConnectInfo) {
	n.l.Lock()
	defer n.l.Unlock()
	n.unsafeStore(key, value)
}

func (n *nodeStore) unsafeStore(key string, value peermessage.ConnectInfo) {
	if 0 == len(n.m) {
		n.a = []peermessage.ConnectInfo{value}
		n.m = map[string]*peermessage.ConnectInfo{
			key: &n.a[0],
		}
	} else {
		v, has := n.m[key]
		if has {
			v.Address = value.Address
			v.Hash = value.Hash
			v.PingTime = value.PingTime
			v.PingScoreBoard = value.PingScoreBoard
		} else {
			n.a = append(n.a, value)
			n.m[key] = &value
		}
	}
	n.db.Update(func(txn *keydb.Tx) error {
		bf := bytes.Buffer{}
		value.WriteTo(&bf)
		bs := bf.Bytes()
		if err := txn.Set([]byte(key), bs, bs); err != nil {
			return errors.WithStack(err)
		}
		return nil
	})
}

// Get returns the value stored in the array for a index
func (n *nodeStore) Get(i int) peermessage.ConnectInfo {
	if i < 0 || i > len(n.a) {
		return peermessage.ConnectInfo{}
	}
	return n.a[i]
}

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (n *nodeStore) Load(key string) (p peermessage.ConnectInfo, has bool) {
	n.l.Lock()
	defer n.l.Unlock()

	v, has := n.m[key]
	if has {
		p = *v
	}
	return
}

// Delete deletes the value for a key.
func (n *nodeStore) Delete(key string) {
	n.l.Lock()
	defer n.l.Unlock()

	delete(n.m, key)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (n *nodeStore) Range(f func(string, peermessage.ConnectInfo) bool) {
	n.l.Lock()
	defer n.l.Unlock()

	for key, value := range n.m {
		if !f(key, *value) {
			break
		}
	}
}

// Len returns the length of this map.
func (n *nodeStore) Len() int {
	n.l.Lock()
	defer n.l.Unlock()

	return len(n.m)
}
