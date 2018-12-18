package cache

import "sync"

// Store uses for temporary storing store from commands until it persistence after '/done'-command
type Store struct {
	rwMu  sync.RWMutex
	store map[string]interface{}
}

func NewStore() *Store {
	return &Store{
		store: make(map[string]interface{}),
	}
}

func (p *Store) Load(key string) interface{} {
	p.rwMu.RLock()
	defer p.rwMu.RUnlock()

	return p.store[key]
}

func (p *Store) Store(key string, value interface{}) {
	p.rwMu.Lock()
	p.store[key] = value
	p.rwMu.Unlock()
}

func (p *Store) Delete(key string) {
	p.rwMu.Lock()
	delete(p.store, key)
	p.rwMu.Unlock()
}
