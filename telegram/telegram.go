package telegram

import (
	"fmt"
	"log"
	"strings"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/incu6us/vote-bot/domain"
	"github.com/incu6us/vote-bot/repository"
	"github.com/incu6us/vote-bot/telegram/models"
	"github.com/incu6us/vote-bot/telegram/polls_cache"
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

type rawCacheInterface interface {
	Load(key string) interface{}
	Store(key string, value interface{})
	Delete(key string)
}

type pollCacheInterface interface {
	Load(key models.UserID) *models.Poll
	Store(key models.UserID, value *models.Poll)
	Delete(key models.UserID)
}

type parseModeType string

const (
	markdownParseMode parseModeType = "markdown"
	htmlParseMode                   = "html"
	noneParseMode                   = ""
)

const (
	isDebug        = true
	parseMode      = markdownParseMode
	maximumAnswers = 3
)

type inlineMessageID string

type Client struct {
	botName         string
	secureUserIDs   []int
	bot             *tgbot.BotAPI
	pollsStore      pollCacheInterface
	store           store
	updatePollCh    chan map[inlineMessageID]*models.UpdatedPoll
	updateMessageCh tgbot.UpdatesChannel
	shutdownCh      chan struct{}
}

func New(cache rawCacheInterface, store store, token, botName string, userIDs ...int) (*Client, error) {
	client := &Client{
		botName: botName,

		secureUserIDs: userIDs,
		pollsStore:    polls_cache.NewPollsStore(cache),
		store:         store,
		updatePollCh:  make(chan map[inlineMessageID]*models.UpdatedPoll),
		shutdownCh:    make(chan struct{}, 1),
	}
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
			for k, values := range updatedPoll.Poll.Votes {
				votes += "\n- " + k + ":\n"
				for _, v := range values {
					votes += "\t\t\t\t" + v + "\n"
				}
			}

			editMsg := tgbot.EditMessageTextConfig{
				BaseEdit: tgbot.BaseEdit{
					InlineMessageID: string(inlineMessageID),
					ReplyMarkup:     preparePollKeyboardMarkup(updatedPoll.Poll),
				},
				Text:      fmt.Sprintf("%s\n---\nLast Vote: %s\nVotes: \n```%s```", escapeURLMarkdownSymbols(updatedPoll.Poll.Subject), updatedPoll.Voter, votes),
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
					if err := c.cmdHelp(update.Message.Chat.ID); err != nil {
						log.Printf("command help: %s\n", err)
					}
				case "cancel":
					if err := c.cmdCancel(update.Message.Chat.ID, update.Message.From.ID); err != nil {
						log.Printf("command cancel: %s\n", err)
					}
				case "done":
					if err := c.cmdDone(update.Message.Chat.ID, update.Message.From.ID); err != nil {
						log.Printf("command done: %s\n", err)
					}
				case "newpoll":
					if err := c.cmdNewPoll(update.Message.Chat.ID, update.Message.From.ID, update.Message.From.String()); err != nil {
						log.Printf("command newpoll: %s\n", err)
					}
				default:
					msg := tgbot.NewMessage(update.Message.Chat.ID, "Bad command")
					if _, err := c.bot.Send(msg); err != nil {
						log.Println("send message error")
					}
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

				if prestoredPoll := c.pollsStore.Load(models.UserID(update.Message.From.ID)); prestoredPoll != nil {
					if prestoredPoll.PollName == "" {
						c.pollsStore.Store(models.UserID(update.Message.From.ID), &models.Poll{PollName: update.Message.Text, Items: []string{}, Owner: getOwner(update.Message.From.ID, update.Message.From.String())})
						c.bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "put items"))
						continue
					}

					if len(prestoredPoll.Items) == maximumAnswers {
						msg := tgbot.NewMessage(update.Message.Chat.ID, "Maximum 3 items could be placed! Use:\n- `/done` - to complete the poll creation;\n- `/cancel` - to cancel the poll creation")
						msg.ParseMode = string(parseMode)
						c.bot.Send(msg)
						continue
					}

					prestoredPoll.Items = append(prestoredPoll.Items, update.Message.Text)
					c.pollsStore.Store(models.UserID(update.Message.From.ID), &models.Poll{PollName: prestoredPoll.PollName, Items: prestoredPoll.Items, Owner: getOwner(update.Message.From.ID, update.Message.From.String())})
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

	c.updatePollCh <- map[inlineMessageID]*models.UpdatedPoll{
		inlineMessageID(callback.InlineMessageID): {
			Voter: callback.From.String(),
			Poll:  poll,
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
