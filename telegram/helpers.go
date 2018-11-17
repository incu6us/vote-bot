package telegram

import (
	"encoding/json"
	"vote-bot/domain"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

// TODO: move to models.go
type callbackData struct {
	PollName string `json:"poll_name"`
	Vote     string `json:"vote"`
}

func preparePollArticle(poll *domain.Poll) tgbot.InlineQueryResultArticle {
	keyboard := new(tgbot.InlineKeyboardMarkup)
	var row []tgbot.InlineKeyboardButton
	for _, item := range poll.Items {
		btn := tgbot.NewInlineKeyboardButtonData(item, prepareCallbackData(poll.Subject, item))
		row = append(row, btn)
	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	resultArticleMarkdown := tgbot.NewInlineQueryResultArticleMarkdown(poll.Subject, poll.Subject, poll.Subject)
	resultArticleMarkdown.ReplyMarkup = keyboard

	return resultArticleMarkdown
}

// TODO: add FFJSON
func prepareCallbackData(pollName, vote string) string {
	data, _ := json.Marshal(callbackData{PollName: pollName, Vote: vote})
	return string(data)
}
