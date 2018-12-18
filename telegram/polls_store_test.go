package telegram

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/incu6us/vote-bot/cache"
)

func Test_pollsStore_Store_Load(t *testing.T) {
	type fields struct {
		tmpStore pollsStoreInterface
	}
	type want struct {
		userID int
		poll   *poll
	}
	type args struct {
		key  userID
		poll *poll
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
				key  userID
				poll *poll
			}{key: 1234, poll: &poll{pollName: "test poll", owner: "me", items: []string{"first item"}}},
			want: struct {
				userID int
				poll   *poll
			}{userID: 1234, poll: &poll{pollName: "test poll", owner: "me", items: []string{"first item"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &pollsStore{
				store: tt.fields.tmpStore,
			}
			p.Store(tt.args.key, tt.args.poll)
			assert.Equal(t, p.Load(userID(tt.want.userID)), tt.want.poll)
		})
	}
}

func Test_pollsStore_Delete(t *testing.T) {
	type fields struct {
		tmpStore pollsStoreInterface
	}
	type prestoredVal struct {
		userID userID
		poll   *poll
	}
	type args struct {
		key userID
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		prestoredVal prestoredVal
	}{
		{
			name:   "success",
			args:   struct{ key userID }{key: userID(1234)},
			fields: struct{ tmpStore pollsStoreInterface }{tmpStore: cache.NewStore()},
			prestoredVal: struct {
				userID userID
				poll   *poll
			}{userID: userID(1234), poll: &poll{pollName: "test poll", owner: "me", items: []string{"first item"}}},
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
