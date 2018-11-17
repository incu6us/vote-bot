package telegram

import (
	"fmt"
	"log"
	"strings"
	"vote-bot/domain"

	"github.com/pkg/errors"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

type store interface {
	GetPolls() ([]*domain.Poll, error)
	GetPoll(pollName string) (*domain.Poll, error)
	CreatePoll(pollName, owner string, items []string) error
	DeletePoll(pollName, owner string) error
	UpdatePollIsPublished(pollName, owner string, isPublished bool) error
	UpdatePollItems(pollName, owner string, items []string) error
}

const debug = true

var (
	polls = make(map[int64]*poll, 100)
)

type Client struct {
	secureChatIDs []int64
	bot           *tgbot.BotAPI
	store         store
}

func Run(token string, chatIDs []int64, store store) error {
	client := &Client{secureChatIDs: chatIDs, store: store}
	if err := client.init(token); err != nil {
		return err
	}

	return nil
}

func (c Client) init(token string) error {
	var err error
	c.bot, err = tgbot.NewBotAPI(token)
	if err != nil {
		return errors.Wrap(err, "telegram bot initialization failed")
	}

	c.bot.Debug = debug
	log.Printf("Authorized on account %s", c.bot.Self.UserName)

	updateConfig := tgbot.NewUpdate(0)
	updateConfig.Timeout = 60

	updateCh, err := c.bot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Printf("get updates failed: %s", err)
		return errors.Wrap(err, "get updates failed")
	}

	for update := range updateCh {
		if update.Message == nil {
			continue
		}

		if update.Message.Chat != nil && !c.chatHasAccess(update.Message.Chat.ID) {
			c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, fmt.Sprintf("You have no access to the bot with chatID: %d", update.Message.Chat.ID)))
			continue
		}

		go func(update tgbot.Update) {
			if update.Message.IsCommand() {
				switch strings.ToLower(update.Message.Command()) {
				case "help":
					msg := tgbot.NewMessage(update.Message.Chat.ID, "")
					msg.ParseMode = "markdown"
					msg.Text = "use this command for help"
					c.bot.Send(msg)
					return
				case "cancel":
					delete(polls, update.Message.Chat.ID)
					c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "canceled"))
					return
				case "done":
					poll, ok := polls[update.Message.Chat.ID]
					if !ok {
						c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "no such poll"))
						return
					}

					defer delete(polls, update.Message.Chat.ID)

					if poll.pollName == "" || len(poll.items) == 0 {
						c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "poll name and items should be set"))
						return
					}

					if err := c.store.CreatePoll(poll.pollName, poll.owner, poll.items); err != nil {
						c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, fmt.Sprintf("poll creation error: %s", err)))
						return
					}

					c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "poll created"))
					return
				case "newpoll":
					if _, ok := polls[update.Message.Chat.ID]; !ok {
						polls[update.Message.Chat.ID] = &poll{owner: fmt.Sprintf("%d-%s %s", update.Message.From.ID, update.Message.From.FirstName, update.Message.From.LastName)}
					}
					msg := tgbot.NewMessage(update.Message.Chat.ID, "enter a poll name")
					c.bot.Send(msg)
					return
				default:
					msg := tgbot.NewMessage(update.Message.Chat.ID, "bad command")
					c.bot.Send(msg)
					return
				}
			} else {
				if strings.TrimSpace(update.Message.Text) == "" {
					c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "message is empty"))
					return
				}

				if _, ok := polls[update.Message.Chat.ID]; ok {
					if polls[update.Message.Chat.ID].pollName == "" {
						polls[update.Message.Chat.ID].pollName = update.Message.Text
						c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "put items"))
						return
					}

					polls[update.Message.Chat.ID].items = append(polls[update.Message.Chat.ID].items, update.Message.Text)
					msg := tgbot.NewMessage(update.Message.Chat.ID, "- put items;\n- `/done` - to complete the poll creation;\n- `/cancel` - to cancel the poll creation")
					msg.ParseMode = "markdown"
					c.bot.Send(msg)
				}
			}
		}(update)
	}

	return nil
}

func (c Client) chatHasAccess(chatID int64) bool {
	for _, secureChatID := range c.secureChatIDs {
		if secureChatID == chatID {
			return true
		}
	}

	return false
}
