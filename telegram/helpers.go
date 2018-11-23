package telegram

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/incu6us/vote-bot/domain"
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
	resultArticleMarkdown := tgbot.NewInlineQueryResultArticleMarkdown(id, subject, escapeURLMarkdownSymbols(poll.Subject))
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

func escapeURLMarkdownSymbols(msg string) string {
	if !strings.Contains(msg, "http://") && !strings.Contains(msg, "https://") {
		return msg
	}

	escapeChars := []string{"_", "*"}
	separator := " "
	httpSeparator := "http"

	words := strings.Split(msg, separator)
	for k, word := range words {
		if !strings.Contains(word, "http://") && !strings.Contains(word, "https://") {
			continue
		}
		httpAddrSlices := strings.Split(word, httpSeparator)
		for i := 1; i < len(httpAddrSlices[0:]); i++ {
			if !strings.HasPrefix(httpAddrSlices[i], "://") && !strings.HasPrefix(httpAddrSlices[i], "s://") {
				continue
			}
			for _, escapeChar := range escapeChars {
				httpAddrSlices[i] = strings.Replace(httpAddrSlices[i], escapeChar, "\\"+escapeChar, -1)
			}
			words[k] = strings.Join(httpAddrSlices, httpSeparator)
		}
	}

	return strings.Join(words, separator)
}
