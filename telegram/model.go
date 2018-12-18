package telegram

import (
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
