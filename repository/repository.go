package repository

import (
	"sync"
	"vote-bot/repository/internal/dynamo"

	"github.com/pkg/errors"
)

type Repository struct {
	sync.Mutex
	db *dynamo.DB
}

func New(region, tableName string) (*Repository, error) {
	db, err := dynamo.New(region, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "create repository failed")
	}

	return &Repository{db: db}, nil
}
