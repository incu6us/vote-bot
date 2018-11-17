package telegram

import (
	"fmt"
	"log"
	"strings"
	"vote-bot/domain"
	"vote-bot/repository"

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
	botName       string
	secureChatIDs []int64
	bot           *tgbot.BotAPI
	store         store
}

func Run(token, botName string, chatIDs []int64, store store) error {
	client := &Client{botName: botName, secureChatIDs: chatIDs, store: store}
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
		if update.InlineQuery != nil {
			if err := c.processInlineRequest(update.InlineQuery); err != nil {
				log.Printf("proccess inline query failed: %s", err)
			}
		}

		if update.Message == nil {
			continue
		}

		if update.Message.Chat != nil && !c.chatHasAccess(update.Message.Chat.ID) {
			c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, fmt.Sprintf("You have no access to the bot with chatID: %d", update.Message.Chat.ID)))
			continue
		}

		if update.Message.IsCommand() {
			switch strings.ToLower(update.Message.Command()) {
			case "help":
				msg := tgbot.NewMessage(update.Message.Chat.ID, "")
				msg.ParseMode = "markdown"
				msg.Text = "use this command for help"
				c.bot.Send(msg)
				continue
			case "cancel":
				delete(polls, update.Message.Chat.ID)
				c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "canceled"))
				continue
			case "done":
				poll, ok := polls[update.Message.Chat.ID]
				if !ok {
					c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "no such poll"))
					continue
				}

				if poll.pollName == "" || len(poll.items) == 0 {
					delete(polls, update.Message.Chat.ID)
					c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "poll name and items should be set. Try again to create a new poll"))
					continue
				}

				if err := c.store.CreatePoll(poll.pollName, poll.owner, poll.items); err != nil {
					delete(polls, update.Message.Chat.ID)
					c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, fmt.Sprintf("poll creation error: %s", err)))
					continue
				}

				delete(polls, update.Message.Chat.ID)

				msg := tgbot.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Poll created. Use `@%s %s`", c.botName, poll.pollName))
				msg.ParseMode = "markdown"
				c.bot.Send(msg)
				continue
			case "newpoll":
				if _, ok := polls[update.Message.Chat.ID]; !ok {
					polls[update.Message.Chat.ID] = &poll{owner: fmt.Sprintf("%d-%s %s", update.Message.From.ID, update.Message.From.FirstName, update.Message.From.LastName)}
				}
				msg := tgbot.NewMessage(update.Message.Chat.ID, "enter a poll name")
				c.bot.Send(msg)
				continue
			default:
				msg := tgbot.NewMessage(update.Message.Chat.ID, "bad command")
				c.bot.Send(msg)
				continue
			}
		} else {
			if strings.TrimSpace(update.Message.Text) == "" {
				c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "message is empty"))
				continue
			}

			if _, ok := polls[update.Message.Chat.ID]; ok {
				if polls[update.Message.Chat.ID].pollName == "" {
					polls[update.Message.Chat.ID].pollName = update.Message.Text
					c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "put items"))
					continue
				}

				polls[update.Message.Chat.ID].items = append(polls[update.Message.Chat.ID].items, update.Message.Text)
				msg := tgbot.NewMessage(update.Message.Chat.ID, "- put items;\n- `/done` - to complete the poll creation;\n- `/cancel` - to cancel the poll creation")
				msg.ParseMode = "markdown"
				c.bot.Send(msg)
			}
		}
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

func (c Client) processInlineRequest(inline *tgbot.InlineQuery) error {
	log.Printf("inline: %+v", inline)

	resultArticlesMarkdown := make([]interface{}, 0, 10)

	if len(inline.Query) <= 3 {
		polls, err := c.store.GetPolls()
		if err != nil {
			if err == repository.ErrPollIsNotFound {
				return nil
			}

			return errors.Wrap(err, "get poll error")
		}

		for _, poll := range polls {
			resultArticleMarkdown := tgbot.NewInlineQueryResultArticleMarkdown(poll.Subject, poll.Subject, "sgvsetgwtfgew")
			resultArticlesMarkdown = append(resultArticlesMarkdown, resultArticleMarkdown)
		}
	} else {
		poll, err := c.store.GetPoll(inline.Query)
		if err != nil {
			if err == repository.ErrPollIsNotFound {
				return nil
			}

			return errors.Wrap(err, "get poll error")
		}

		resultArticleMarkdown := tgbot.NewInlineQueryResultArticleMarkdown(poll.Subject, poll.Subject, "sgvsetgwtfgew")
		resultArticlesMarkdown = append(resultArticlesMarkdown, resultArticleMarkdown)
	}

	inlineConfig := tgbot.InlineConfig{
		InlineQueryID: inline.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       resultArticlesMarkdown,
	}

	_, err := c.bot.AnswerInlineQuery(inlineConfig)
	if err != nil {
		return errors.Wrap(err, "answer inline error")
	}

	return nil
}
