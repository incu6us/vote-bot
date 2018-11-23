package telegram

import (
	"fmt"
	"log"
	"strings"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/incu6us/vote-bot/domain"
	"github.com/incu6us/vote-bot/repository"
	"github.com/pkg/errors"
)

type store interface {
	GetPolls() ([]*domain.Poll, error)
	GetPoll(pollName string) (*domain.Poll, error)
	GetPollBeginsWith(pollName string) (*domain.Poll, error)
	CreatePoll(pollName, owner string, items []string) error
	DeletePoll(pollName, owner string) error
	UpdatePollIsPublished(pollName, owner string, isPublished bool) error
	UpdatePollItems(pollName, owner string, items []string) error
	UpdateVote(createdAt int64, item, user string) (*domain.Poll, error)
}

const (
	isDebug        = true
	parseMode      = markdownParseMode
	maximumAnswers = 3
)

var (
	polls = newPollsMemStore()
)

type inlineMessageID string

type Client struct {
	botName         string
	secureUserIDs   []int
	bot             *tgbot.BotAPI
	store           store
	updatePollCh    chan map[inlineMessageID]*updatedPoll
	updateMessageCh tgbot.UpdatesChannel
	shutdownCh      chan struct{}
}

func New(token, botName string, userIDs []int, store store) (*Client, error) {
	client := &Client{botName: botName, secureUserIDs: userIDs, store: store, updatePollCh: make(chan map[inlineMessageID]*updatedPoll), shutdownCh: make(chan struct{}, 1)}
	if err := client.login(token); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) Run() error {
	updateConfig := tgbot.NewUpdate(0)
	updateConfig.Timeout = 60

	var err error
	c.updateMessageCh, err = c.bot.GetUpdatesChan(updateConfig)
	if err != nil {
		return errors.Wrap(err, "get updates failed")
	}

	go c.updatePollAnswers()
	c.messageListen()

	return nil
}

func (c *Client) login(token string) error {
	var err error
	c.bot, err = tgbot.NewBotAPI(token)
	if err != nil {
		return errors.Wrap(err, "telegram bot initialization failed")
	}

	c.bot.Debug = isDebug
	log.Printf("Authorized on account %s", c.bot.Self.UserName)

	return nil
}

func (c *Client) updatePollAnswers() {
	for update := range c.updatePollCh {
		for inlineMessageID, updatedPoll := range update {
			var votes string
			for k, values := range updatedPoll.poll.Votes {
				votes += "\n- " + k + ":\n"
				for _, v := range values {
					votes += "\t\t\t\t" + v + "\n"
				}
			}

			editMsg := tgbot.EditMessageTextConfig{
				BaseEdit: tgbot.BaseEdit{
					InlineMessageID: string(inlineMessageID),
					ReplyMarkup:     preparePollKeyboardMarkup(updatedPoll.poll),
				},
				Text:      fmt.Sprintf("%s\n---\nLast Vote: %s\nVotes: \n```%s```", escapeURLMarkdownSymbols(updatedPoll.poll.Subject), updatedPoll.voter, votes),
				ParseMode: string(parseMode),
			}

			if _, err := c.bot.Send(editMsg); err != nil {
				log.Printf("update message error: %s", err)
			}
		}
	}
}

func (c *Client) messageListen() {
	for {
		select {
		case <-c.shutdownCh:
			return
		case update := <-c.updateMessageCh:
			if update.Message == nil && update.CallbackQuery != nil {
				if err := c.processPollAnswer(update.CallbackQuery); err != nil {
					log.Printf("prccess callback error: %s", err)
					continue
				}
			}

			if update.Message == nil && update.InlineQuery != nil {
				if !c.userHasAccess(update.InlineQuery.From.ID) {
					c.bot.Send(tgbot.NewMessage(int64(update.InlineQuery.From.ID), msgYouHaveNoAccess(int64(update.InlineQuery.From.ID))))
					continue
				}

				if err := c.postPoll(update.InlineQuery); err != nil {
					log.Printf("proccess inline query failed: %s", err)
				}
			}

			if update.Message == nil {
				continue
			}

			if update.Message.Chat != nil && !c.userHasAccess(update.Message.From.ID) {
				c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, msgYouHaveNoAccess(update.Message.Chat.ID)))
				continue
			}

			if update.Message.IsCommand() {
				switch strings.ToLower(update.Message.Command()) {
				case "help":
					msg := tgbot.NewMessage(update.Message.Chat.ID, "")
					msg.ParseMode = string(parseMode)
					msg.Text = "use this command for help"
					c.bot.Send(msg)
					continue
				case "cancel":
					polls.Delete(userID(update.Message.From.ID))
					c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "Canceled"))
					continue
				case "done":
					poll := polls.Load(userID(update.Message.From.ID))
					if poll == nil {
						c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "No such poll"))
						continue
					}

					if poll.pollName == "" || len(poll.items) == 0 {
						polls.Delete(userID(update.Message.From.ID))
						c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "Poll name and items should be set. Try again to create a new poll"))
						continue
					}

					if err := c.store.CreatePoll(poll.pollName, poll.owner, poll.items); err != nil {
						polls.Delete(userID(update.Message.From.ID))
						c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Poll creation error: %s", err)))
						continue
					}

					polls.Delete(userID(update.Message.From.ID))

					msg := tgbot.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Use `share button` or put the next lines into your group: `@%s %s`", c.botName, poll.pollName))
					msg.ParseMode = string(parseMode)
					msg.ReplyMarkup = &tgbot.InlineKeyboardMarkup{
						InlineKeyboard: [][]tgbot.InlineKeyboardButton{
							{
								{
									Text:              "Share with group",
									SwitchInlineQuery: stringToPtr(poll.pollName),
								},
							},
						},
					}
					c.bot.Send(msg)
					continue
				case "newpoll":
					if prestoredPoll := polls.Load(userID(update.Message.From.ID)); prestoredPoll == nil {
						polls.Store(userID(update.Message.From.ID), &poll{owner: getOwner(update.Message.From.ID, update.Message.From.String())})
					}
					msg := tgbot.NewMessage(update.Message.Chat.ID, "Enter a poll name")
					c.bot.Send(msg)
					continue
				default:
					msg := tgbot.NewMessage(update.Message.Chat.ID, "Bad command")
					c.bot.Send(msg)
					continue
				}
			} else {
				log.Printf("MESSAGE %+v", update.Message)
				if update.Message.NewChatMembers != nil || update.Message.LeftChatMember != nil {
					continue
				}
				if strings.TrimSpace(update.Message.Text) == "" {
					c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "message is empty"))
					continue
				}

				if prestoredPoll := polls.Load(userID(update.Message.From.ID)); prestoredPoll != nil {
					if prestoredPoll.pollName == "" {
						polls.Store(userID(update.Message.From.ID), &poll{pollName: update.Message.Text, items: []string{}, owner: getOwner(update.Message.From.ID, update.Message.From.String())})
						c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "put items"))
						continue
					}

					if len(prestoredPoll.items) == maximumAnswers {
						msg := tgbot.NewMessage(update.Message.Chat.ID, "Maximum 3 items could be placed! Use:\n- `/done` - to complete the poll creation;\n- `/cancel` - to cancel the poll creation")
						msg.ParseMode = string(parseMode)
						c.bot.Send(msg)
						continue
					}

					prestoredPoll.items = append(prestoredPoll.items, update.Message.Text)
					polls.Store(userID(update.Message.From.ID), &poll{pollName: prestoredPoll.pollName, items: prestoredPoll.items, owner: getOwner(update.Message.From.ID, update.Message.From.String())})
					msg := tgbot.NewMessage(update.Message.Chat.ID, "- put items;\n- `/done` - to complete the poll creation;\n- `/cancel` - to cancel the poll creation")
					msg.ParseMode = string(parseMode)
					c.bot.Send(msg)
				}
			}
		}
	}
}

func (c Client) userHasAccess(userID int) bool {
	for _, securedUserID := range c.secureUserIDs {
		if securedUserID == userID {
			return true
		}
	}

	return false
}

func (c Client) postPoll(inline *tgbot.InlineQuery) error {
	if len(inline.Query) <= 3 {
		return nil
	}

	poll, err := c.store.GetPollBeginsWith(inline.Query)
	if err != nil {
		if err == repository.ErrPollIsNotFound {
			return nil
		}

		return errors.Wrap(err, "get poll error")
	}

	resultArticlesMarkdown := []interface{}{
		preparePollArticle(poll),
	}

	inlineConfig := tgbot.InlineConfig{
		InlineQueryID: inline.ID,
		IsPersonal:    false,
		CacheTime:     0,
		Results:       resultArticlesMarkdown,
	}

	_, err = c.bot.AnswerInlineQuery(inlineConfig)
	if err != nil {
		return errors.Wrap(err, "answer inline error")
	}

	return nil
}

func (c Client) processPollAnswer(callback *tgbot.CallbackQuery) error {
	callbackData, err := serializeCallbackData(callback.Data)
	if err != nil {
		return errors.Wrap(err, "get callback data error")
	}

	poll, err := c.store.UpdateVote(callbackData.CreatedAt, callbackData.Vote, callback.From.String())
	if err != nil {
		return errors.Wrap(err, "update vote failed")
	}

	log.Printf("POLL: %+v", poll)

	callbackConfig := tgbot.CallbackConfig{
		CallbackQueryID: callback.ID,
		Text:            fmt.Sprintf("Vote '%s' accepted", callbackData.Vote),
		ShowAlert:       false,
		URL:             "",
		CacheTime:       0,
	}

	if _, err := c.bot.AnswerCallbackQuery(callbackConfig); err != nil {
		return errors.Wrap(err, "answer callback error")
	}

	c.updatePollCh <- map[inlineMessageID]*updatedPoll{
		inlineMessageID(callback.InlineMessageID): {
			voter: callback.From.String(),
			poll:  poll,
		},
	}

	return nil
}

func (c *Client) Close() error {
	c.shutdownCh <- struct{}{}
	c.bot.StopReceivingUpdates()
	close(c.updatePollCh)
	close(c.shutdownCh)

	return nil
}
