package closer

import (
	"log"
	"sync"
)

// Closer is Closer inferface
type Closer interface {
	Close()
}

// Manager handles closers
type Manager struct {
	isClosed bool
	Names    []string
	Closers  []Closer
	wg       sync.WaitGroup
}

// NewManager returns a Manager
func NewManager() *Manager {
	cm := &Manager{
		Names:   []string{},
		Closers: []Closer{},
	}
	cm.wg.Add(1)
	return cm
}

// IsClosed returns it is closed or not
func (cm *Manager) IsClosed() bool {
	return cm.isClosed
}

// RemoveAll removes all closers
func (cm *Manager) RemoveAll() {
	cm.Names = []string{}
	cm.Closers = []Closer{}
}

// Add adds a closer with a name
func (cm *Manager) Add(Name string, c Closer) {
	cm.Names = append(cm.Names, Name)
	cm.Closers = append(cm.Closers, c)
}

// CloseAll closers all closers
func (cm *Manager) CloseAll() {
	if !cm.isClosed {
		cm.isClosed = true
		for i, c := range cm.Closers {
			log.Println("Close", cm.Names[i])
			c.Close()
		}
		cm.wg.Done()
	}
}

// Wait waits close all
func (cm *Manager) Wait() {
	cm.wg.Wait()
}
