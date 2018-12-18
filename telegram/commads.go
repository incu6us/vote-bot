package telegram

import (
	"fmt"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/incu6us/vote-bot/telegram/models"
	"github.com/pkg/errors"
)

const (
	sendMessageErrorString = "send message error"
)

func (c Client) cmdHelp(chatID int64) error {
	msg := tgbot.NewMessage(chatID, "")
	msg.ParseMode = string(parseMode)
	msg.Text = "use this command for help"

	if _, err := c.bot.Send(msg); err != nil {
		return errors.Wrap(err, sendMessageErrorString)
	}

	return nil
}

func (c *Client) cmdCancel(chatID int64, userID int) error {
	c.pollsStore.Delete(models.UserID(userID))
	if _, err := c.bot.Send(tgbot.NewMessage(chatID, "Canceled")); err != nil {
		return errors.Wrap(err, "cmd cancel error")
	}

	return nil
}

func (c *Client) cmdDone(chatID int64, userID int) error {
	poll := c.pollsStore.Load(models.UserID(userID))
	if poll == nil {
		if _, err := c.bot.Send(tgbot.NewMessage(chatID, "No such poll")); err != nil {
			return errors.Wrap(err, sendMessageErrorString)
		}

		return nil
	}

	if poll.PollName == "" || len(poll.Items) == 0 {
		c.pollsStore.Delete(models.UserID(userID))
		if _, err := c.bot.Send(tgbot.NewMessage(chatID, "Poll name and items should be set. Try again to create a new poll")); err != nil {
			return errors.Wrap(err, sendMessageErrorString)
		}

		return nil
	}

	if err := c.store.CreatePoll(poll.PollName, poll.Owner, poll.Items); err != nil {
		c.pollsStore.Delete(models.UserID(userID))
		if _, err := c.bot.Send(tgbot.NewMessage(chatID, fmt.Sprintf("Poll creation error: %s", err))); err != nil {
			return errors.Wrap(err, sendMessageErrorString)
		}
		return nil
	}

	c.pollsStore.Delete(models.UserID(userID))

	msg := tgbot.NewMessage(chatID, fmt.Sprintf("Use `share button` or put the next lines into your group: `@%s %s`", c.botName, poll.PollName))
	msg.ParseMode = string(parseMode)
	msg.ReplyMarkup = &tgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbot.InlineKeyboardButton{
			{
				{
					Text:              "Share with group",
					SwitchInlineQuery: stringToPtr(poll.PollName),
				},
			},
		},
	}

	if _, err := c.bot.Send(msg); err != nil {
		return errors.Wrap(err, sendMessageErrorString)
	}

	return nil
}

func (c *Client) cmdNewPoll(chatID int64, userID int, fullUserName string) error {
	if prestoredPoll := c.pollsStore.Load(models.UserID(userID)); prestoredPoll == nil {
		c.pollsStore.Store(models.UserID(userID), &models.Poll{Owner: getOwner(userID, fullUserName)})
	}
	msg := tgbot.NewMessage(chatID, "Enter a poll name")
	if _, err := c.bot.Send(msg); err != nil {
		return errors.Wrap(err, sendMessageErrorString)
	}

	return nil
}
