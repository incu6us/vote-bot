package repository

import (
	"sync"

	"vote-bot/repository/internal/dynamo"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/pkg/errors"
)

const (
	pollsAmount = 10
)

var (
	ErrPollIsNotFound = errors.New("poll is not found")
)

type Repository struct {
	mu sync.Mutex
	db *dynamo.DB
}

func New(region, tableName string) (*Repository, error) {
	db, err := dynamo.New(region, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "create repository failed")
	}

	return &Repository{db: db}, nil
}

func (r *Repository) GetPolls() ([]*Poll, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result, err := r.db.GetPolls(pollsAmount)
	if err != nil {
		return nil, errors.Wrap(err, "can't get polls from repository")
	}

	return r.convertMapToPoll(result.Items...)
}

func (r *Repository) GetPoll(pollName string) (*Poll, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, err := r.db.GetPoll(pollName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get a poll by name")
	}

	if item == nil || item.Item == nil || len(item.Item) == 0 {
		return nil, ErrPollIsNotFound
	}

	poll := new(Poll)
	if err := dynamodbattribute.UnmarshalMap(item.Item, poll); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal item")
	}

	return poll, nil
}

func (r Repository) convertMapToPoll(items ...map[string]*dynamodb.AttributeValue) ([]*Poll, error) {
	polls := make([]*Poll, len(items))

	for i, item := range items {
		poll := new(Poll)
		if err := dynamodbattribute.UnmarshalMap(item, poll); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal map for item")
		}

		polls[i] = poll
	}

	return polls, nil
}
