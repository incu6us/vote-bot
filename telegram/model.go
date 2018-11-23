package telegram

import (
	"sync"

	"github.com/incu6us/vote-bot/domain"
)

type parseModeType string

const (
	markdownParseMode parseModeType = "markdown"
	htmlParseMode                   = "html"
	noneParseMode                   = ""
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
	rwMu  sync.RWMutex
	polls map[userID]*poll
}

func newPollsMemStore() *pollsMemStore {
	return &pollsMemStore{
		polls: make(map[userID]*poll),
	}
}

func (p *pollsMemStore) Load(key userID) *poll {
	p.rwMu.RLock()
	defer p.rwMu.RUnlock()

	return p.polls[key]
}

func (p *pollsMemStore) Store(key userID, poll *poll) {
	p.rwMu.Lock()
	defer p.rwMu.Unlock()

	p.polls[key] = poll
}

func (p *pollsMemStore) Delete(key userID) {
	p.rwMu.Lock()
	defer p.rwMu.Unlock()

	delete(p.polls, key)
}
