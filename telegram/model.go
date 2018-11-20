package telegram

import "vote-bot/domain"

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
