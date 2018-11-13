package repository

import (
	"sync"
	"time"

	"vote-bot/repository/internal/dynamo"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/pkg/errors"
)

const (
	pollsAmount = 10
)

var (
	ErrPollIsNotFound   = errors.New("poll is not found")
	ErrPollAlreadyExist = errors.New("poll already exist")
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

func (r *Repository) CreateTable() error {
	return r.db.CreateTable()
}

func (r *Repository) DescribeTable() (string, error) {
	return r.db.DescribeTable()
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

	return r.getPoll(pollName)
}

func (r *Repository) CreatePoll(pollName string, items []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	storedPoll, err := r.getPoll(pollName)
	if err != nil {
		return errors.Wrap(err, "create poll failed")
	}

	if storedPoll != nil {
		return ErrPollAlreadyExist
	}

	poll := &Poll{
		CreatedAt: time.Now().UnixNano(),
		Subject:   pollName,
		Items:     items,
		Kind:      dynamo.PollKind,
	}

	return r.db.CreatePoll(poll)
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

func (r Repository) getPoll(pollName string) (*Poll, error) {
	item, err := r.db.GetPoll(pollName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get a poll by name")
	}

	if item == nil || item.Items == nil || len(item.Items) == 0 {
		return nil, ErrPollIsNotFound
	}

	poll := new(Poll)
	if err := dynamodbattribute.UnmarshalMap(item.Items[0], poll); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal item")
	}

	return poll, nil
}
