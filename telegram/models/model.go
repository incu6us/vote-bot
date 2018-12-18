package models

import (
	"github.com/incu6us/vote-bot/domain"
)

type Poll struct {
	PollName, Owner string
	Items           []string
}

type CallbackData struct {
	CreatedAt int64  `json:"created_at"`
	Vote      string `json:"vote"`
}

type UpdatedPoll struct {
	Voter string
	Poll  *domain.Poll
}

type UserID int
