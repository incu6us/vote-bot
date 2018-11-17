package telegram

import (
	"vote-bot/domain"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

func preparePollArticle(poll *domain.Poll) tgbot.InlineQueryResultArticle {
	keyboard := new(tgbot.InlineKeyboardMarkup)
	var row []tgbot.InlineKeyboardButton
	for _, item := range poll.Items {
		btn := tgbot.NewInlineKeyboardButtonData(item, item)
		row = append(row, btn)
	}
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	resultArticleMarkdown := tgbot.NewInlineQueryResultArticleMarkdown(poll.Subject, poll.Subject, poll.Subject)
	resultArticleMarkdown.ReplyMarkup = keyboard

	return resultArticleMarkdown
}
