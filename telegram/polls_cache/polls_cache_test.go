package polls_cache

import (
	"testing"

	"github.com/incu6us/vote-bot/telegram/models"

	"github.com/stretchr/testify/assert"

	"github.com/incu6us/vote-bot/cache"
)

func Test_pollsStore_Store_Load(t *testing.T) {
	type fields struct {
		tmpStore pollsStoreInterface
	}
	type want struct {
		userID int
		poll   *models.Poll
	}
	type args struct {
		key  models.UserID
		poll *models.Poll
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name:   "success",
			fields: struct{ tmpStore pollsStoreInterface }{tmpStore: cache.NewStore()},
			args: struct {
				key  models.UserID
				poll *models.Poll
			}{key: 1234, poll: &models.Poll{PollName: "test poll", Owner: "me", Items: []string{"first item"}}},
			want: struct {
				userID int
				poll   *models.Poll
			}{userID: 1234, poll: &models.Poll{PollName: "test poll", Owner: "me", Items: []string{"first item"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &pollsStore{
				store: tt.fields.tmpStore,
			}
			p.Store(tt.args.key, tt.args.poll)
			assert.Equal(t, p.Load(models.UserID(tt.want.userID)), tt.want.poll)
		})
	}
}

func Test_pollsStore_Delete(t *testing.T) {
	type fields struct {
		tmpStore pollsStoreInterface
	}
	type prestoredVal struct {
		userID models.UserID
		poll   *models.Poll
	}
	type args struct {
		key models.UserID
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		prestoredVal prestoredVal
	}{
		{
			name:   "success",
			args:   struct{ key models.UserID }{key: models.UserID(1234)},
			fields: struct{ tmpStore pollsStoreInterface }{tmpStore: cache.NewStore()},
			prestoredVal: struct {
				userID models.UserID
				poll   *models.Poll
			}{userID: models.UserID(1234), poll: &models.Poll{PollName: "test poll", Owner: "me", Items: []string{"first item"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &pollsStore{
				store: tt.fields.tmpStore,
			}
			p.Store(tt.prestoredVal.userID, tt.prestoredVal.poll)
			assert.Equal(t, tt.prestoredVal.poll, p.Load(tt.prestoredVal.userID))
			p.Delete(tt.args.key)
			assert.Nil(t, p.Load(tt.prestoredVal.userID))
		})
	}
}
