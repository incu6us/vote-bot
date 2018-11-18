package telegram

import (
	"encoding/json"
	"fmt"
	"vote-bot/domain"

	"github.com/pkg/errors"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

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

func serializeCallbackData(data string) (*callbackData, error) {
	callbackData := new(callbackData)
	if err := json.Unmarshal([]byte(data), callbackData); err != nil {
		return nil, errors.Wrap(err, "serialize callback data error")
	}

	return callbackData, nil
}

func msgYouHaveNoAccess(id int64) string {
	return fmt.Sprintf("You have no access to the bot with userID: %d", id)
}
