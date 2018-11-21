package telegram

import (
	"sync"
	"vote-bot/domain"
)

type poll struct {
	pollName, owner string
	items           []string
}

type callbackData struct {
	CreatedAt int64  `json:"created_at"`
	Vote      string `json:"vote"`
}

type updatedPoll struct {
	voter string
	poll  *domain.Poll
}

type userID int

// pollsMemStore uses for temporary storing polls from commands until it persistence after '/done'-command
type pollsMemStore struct {
	sync.RWMutex
	polls map[userID]*poll
}

func newPollsMemStore() *pollsMemStore {
	return &pollsMemStore{
		polls: make(map[userID]*poll),
	}
}

func (p *pollsMemStore) Load(key userID) *poll {
	p.RLock()
	defer p.RUnlock()

	return p.polls[key]
}

func (p *pollsMemStore) Store(key userID, poll *poll) {
	p.Lock()
	defer p.Unlock()

	p.polls[key] = poll
}

func (p *pollsMemStore) Delete(key userID) {
	p.Lock()
	defer p.Unlock()

	delete(p.polls, key)
}
