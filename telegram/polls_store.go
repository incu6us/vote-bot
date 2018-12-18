package telegram

import "strconv"

type pollsStoreInterface interface {
	Load(key string) interface{}
	Store(key string, value interface{})
	Delete(key string)
}

type pollsStore struct {
	store pollsStoreInterface
}

func newPollsStore(store pollsStoreInterface) *pollsStore {
	return &pollsStore{store: store}
}

func (p pollsStore) Load(key userID) *poll {
	if val := p.store.Load(strconv.Itoa(int(key))); val != nil {
		return val.(*poll)
	}

	return nil
}

func (p *pollsStore) Store(key userID, poll *poll) {
	p.store.Store(strconv.Itoa(int(key)), poll)
}

func (p *pollsStore) Delete(key userID) {
	p.store.Delete(strconv.Itoa(int(key)))
}
