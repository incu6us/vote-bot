package repository

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/incu6us/vote-bot/domain"
	"github.com/incu6us/vote-bot/repository/internal/dynamo"
	"github.com/pkg/errors"
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

func (r *Repository) GetPolls() ([]*domain.Poll, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result, err := r.db.GetPolls()
	if err != nil {
		return nil, errors.Wrap(err, "can't get polls from repository")
	}

	return r.convertMapToPoll(result.Items...)
}

func (r *Repository) GetPoll(pollName string) (*domain.Poll, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.getPoll(strings.TrimSpace(pollName))
}

func (r *Repository) GetPollBeginsWith(pollName string) (*domain.Poll, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, err := r.db.GetPollBeginsWith(strings.TrimSpace(pollName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get a poll by name")
	}

	if item == nil || item.Items == nil || len(item.Items) == 0 {
		return nil, ErrPollIsNotFound
	}

	poll := new(domain.Poll)
	if err := dynamodbattribute.UnmarshalMap(item.Items[0], poll); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal item")
	}

	return poll, nil
}

func (r *Repository) CreatePoll(pollName, owner string, items []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	storedPoll, err := r.getPoll(strings.TrimSpace(pollName))
	if err != nil && errors.Cause(err) != ErrPollIsNotFound {
		return errors.Wrap(err, "create poll failed")
	}

	if storedPoll != nil {
		return ErrPollAlreadyExist
	}

	poll := &domain.Poll{
		CreatedAt: time.Now().UnixNano(),
		Subject:   strings.TrimSpace(pollName),
		Items:     items,
		Votes:     map[string][]string{},
		CreatedBy: owner,
	}

	item, err := dynamodbattribute.MarshalMap(poll)
	if err != nil {
		return errors.Wrap(err, "filed to marshal an item")
	}

	return r.db.CreatePoll(item)
}

func (r *Repository) DeletePoll(pollName, owner string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result, err := r.db.GetPollByOwner(strings.TrimSpace(pollName), owner)
	if err != nil {
		return err
	}

	if len(result.Items) == 0 {
		return ErrPollIsNotFound
	}

	var poll domain.Poll
	if err := dynamodbattribute.UnmarshalMap(result.Items[0], &poll); err != nil {
		return err
	}

	return r.db.DeletePoll(strings.TrimSpace(pollName), poll.CreatedAt)
}

func (r *Repository) UpdatePollIsPublished(pollName, owner string, isPublished bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result, err := r.db.GetPollByOwner(strings.TrimSpace(pollName), owner)
	if err != nil {
		return err
	}

	if len(result.Items) == 0 {
		return ErrPollIsNotFound
	}

	var poll domain.Poll
	if err := dynamodbattribute.UnmarshalMap(result.Items[0], &poll); err != nil {
		return err
	}

	return r.db.UpdateIsPublish(strings.TrimSpace(pollName), poll.CreatedAt, isPublished)
}

func (r *Repository) UpdatePollItems(pollName, owner string, items []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result, err := r.db.GetPollByOwner(strings.TrimSpace(pollName), owner)
	if err != nil {
		return err
	}

	if len(result.Items) == 0 {
		return ErrPollIsNotFound
	}

	var poll domain.Poll
	if err := dynamodbattribute.UnmarshalMap(result.Items[0], &poll); err != nil {
		return err
	}

	return r.db.UpdateItems(strings.TrimSpace(pollName), poll.CreatedAt, items)
}

func (r *Repository) UpdateVote(createdAt int64, item, voter string) (*domain.Poll, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	poll, err := r.getPollByCreatedAt(createdAt)
	if err != nil {
		return nil, errors.Wrap(err, "get poll failed")
	}

	// delete previous vote fo the user
	for item, users := range poll.Votes {
		for i, user := range users {
			if user == voter {
				users = append(users[:i], users[i+1:]...)
				if len(users) > 0 {
					poll.Votes[item] = users
				} else {
					delete(poll.Votes, item)
				}
			}
		}
	}

	// add user to vote item
	if poll.Votes == nil {
		poll.Votes = make(map[string][]string)
	}

	if _, ok := poll.Votes[item]; !ok {
		poll.Votes[item] = []string{voter}
	} else {
		poll.Votes[item] = append(poll.Votes[item], voter)
	}

	voteAttributes, err := dynamodbattribute.MarshalMap(poll.Votes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal votes")
	}

	if err := r.db.UpdateVotes(poll.Subject, poll.CreatedAt, voteAttributes); err != nil {
		return nil, errors.Wrap(err, "failed to update vote in database")
	}

	return poll, nil
}

func (r Repository) convertMapToPoll(items ...map[string]*dynamodb.AttributeValue) ([]*domain.Poll, error) {
	polls := make([]*domain.Poll, len(items))

	for i, item := range items {
		poll := new(domain.Poll)
		if err := dynamodbattribute.UnmarshalMap(item, poll); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal map for item")
		}

		polls[i] = poll
	}

	return polls, nil
}

func (r Repository) getPollByCreatedAt(createdAt int64) (*domain.Poll, error) {
	items, err := r.db.GetPollByCreatedAt(createdAt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get a poll by created_at field")
	}

	log.Printf("ITEM: %+v", items)
	if items == nil || items.Items == nil || len(items.Items) == 0 {
		return nil, ErrPollIsNotFound
	}

	poll := new(domain.Poll)
	if err := dynamodbattribute.UnmarshalMap(items.Items[0], poll); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal items")
	}

	return poll, nil
}

func (r Repository) getPoll(pollName string) (*domain.Poll, error) {
	item, err := r.db.GetPoll(pollName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get a poll by name")
	}

	if item == nil || item.Items == nil || len(item.Items) == 0 {
		return nil, ErrPollIsNotFound
	}

	poll := new(domain.Poll)
	if err := dynamodbattribute.UnmarshalMap(item.Items[0], poll); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal item")
	}

	return poll, nil
}
