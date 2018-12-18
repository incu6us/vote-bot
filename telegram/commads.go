package telegram

import (
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

func (c Client) sendHelp(chatID int64) error {
	msg := tgbot.NewMessage(chatID, "")
	msg.ParseMode = string(parseMode)
	msg.Text = "use this command for help"

	_, err := c.bot.Send(msg)
	return errors.Wrap(err, "send help error")
}
