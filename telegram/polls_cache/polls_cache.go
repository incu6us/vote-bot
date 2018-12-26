package polls_cache

import (
	"strconv"

	"github.com/incu6us/vote-bot/telegram/models"
)

type pollsStoreInterface interface {
	Load(key string) interface{}
	Store(key string, value interface{})
	Delete(key string)
}

type pollsStore struct {
	store pollsStoreInterface
}

func NewPollsStore(store pollsStoreInterface) *pollsStore {
	return &pollsStore{store: store}
}

func (p pollsStore) Load(key models.UserID) *models.Poll {
	if val := p.store.Load(strconv.Itoa(int(key))); val != nil {
		return val.(*models.Poll)
	}

	return nil
}

func (p *pollsStore) Store(key models.UserID, poll *models.Poll) {
	p.store.Store(strconv.Itoa(int(key)), poll)
}

func (p *pollsStore) Delete(key models.UserID) {
	p.store.Delete(strconv.Itoa(int(key)))
}
