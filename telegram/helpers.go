package telegram

import (
	"encoding/json"
	"fmt"
	"strconv"
	"vote-bot/domain"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

const (
	inlineButtonLength = 32
)

func preparePollArticle(poll *domain.Poll) tgbot.InlineQueryResultArticle {
	id := strconv.FormatInt(poll.CreatedAt, 10)
	subject := poll.Subject
	if len(subject) >= inlineButtonLength {
		subject = poll.Subject[:inlineButtonLength] + "..."
	}
	resultArticleMarkdown := tgbot.NewInlineQueryResultArticleMarkdown(id, subject, poll.Subject)
	resultArticleMarkdown.ReplyMarkup = preparePollKeyboardMarkup(poll)

	return resultArticleMarkdown
}

func preparePollKeyboardMarkup(poll *domain.Poll) *tgbot.InlineKeyboardMarkup {
	keyboard := new(tgbot.InlineKeyboardMarkup)
	var row []tgbot.InlineKeyboardButton
	for _, item := range poll.Items {
		btn := tgbot.NewInlineKeyboardButtonData(item, prepareCallbackData(poll.CreatedAt, item))
		row = append(row, btn)
	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	return keyboard
}

// TODO: add FFJSON
func prepareCallbackData(createdAt int64, vote string) string {
	data, _ := json.Marshal(callbackData{CreatedAt: createdAt, Vote: vote})
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

func getOwner(id int, name string) string {
	return fmt.Sprintf("(%d) %s", id, name)
}
