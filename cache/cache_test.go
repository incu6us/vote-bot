package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStore_Store_Load(t *testing.T) {
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: struct {
				key   string
				value interface{}
			}{key: "test", value: 123},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewStore()
			p.Store(tt.args.key, tt.args.value)
			assert.Equal(t, tt.args.value, p.Load(tt.args.key))
		})
	}
}

func TestStore_Delete(t *testing.T) {
	type storedData struct {
		id   string
		data interface{}
	}
	type args struct {
		id string
	}
	tests := []struct {
		name       string
		args       args
		storedData storedData
	}{
		{
			name: "success",
			args: struct{ id string }{id: "key"},
			storedData: struct {
				id   string
				data interface{}
			}{id: "key", data: 123},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewStore()
			p.Store(tt.storedData.id, tt.storedData.data)
			assert.Equal(t, tt.storedData.data, p.Load(tt.storedData.id))
			p.Delete(tt.args.id)
			assert.Nil(t, p.Load(tt.storedData.id))
		})
	}
}
