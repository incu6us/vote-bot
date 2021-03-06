package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/incu6us/vote-bot/cache"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	cfg "github.com/incu6us/vote-bot/config"
	"github.com/incu6us/vote-bot/repository"
	"github.com/incu6us/vote-bot/telegram"
	"github.com/pkg/errors"
)

const (
	cfgPrefix = "VB"
	cfgFile   = "config.json"
)

func config() (*cfg.Config, error) {
	return cfg.New(cfgPrefix, cfgFile, cfg.JSONConfigType)
}

func main() {
	cfg, err := config()
	if err != nil {
		log.Printf("failed to read config file: %s", err)
		return
	}

	region := cfg.GetString("region")
	if region == "" {
		log.Printf("region is not set")
		return
	}

	tableName := cfg.GetString("dynamo.table")
	if tableName == "" {
		log.Printf("dynamo table is not set")
		return
	}

	telegramToken := cfg.GetString("telegram.token")
	if telegramToken == "" {
		log.Printf("telegram token is not set")
		return
	}

	botName := cfg.GetString("telegram.bot_name")
	if botName == "" {
		log.Printf("telegram bot name is not set")
		return
	}

	userIDSlice := cfg.Get("telegram.user_ids").([]interface{})
	if len(userIDSlice) == 0 {
		log.Printf("telegram userIDs is not set")
		return
	}

	userIDs := make([]int, 0)
	for _, userID := range userIDSlice {
		for _, id := range userID.(map[string]interface{}) {
			userIDs = append(userIDs, int(id.(float64)))
		}
	}

	repo, err := repository.New(region, tableName)
	if err != nil {
		log.Printf("failed to initiate repository: %s", err)
		return
	}

	if _, err := repo.DescribeTable(); err != nil {
		if awsErr, ok := errors.Cause(err).(awserr.Error); ok {
			switch awsErr.Code() {
			case dynamodb.ErrCodeResourceNotFoundException:
				if err1 := repo.CreateTable(); err1 != nil {
					log.Printf("create table error: %s", err1)
				}
				log.Println("table created")
			}
		} else {
			log.Println(err)
			return
		}
	}

	bot, err := telegram.New(cache.NewStore(), repo, telegramToken, botName, userIDs...)
	if err != nil {
		log.Printf("bot creation error: %s\n", err)
		return
	}

	go shutdown(bot)

	if err := bot.Run(); err != nil {
		log.Printf("bot start failed: %s\n", err)
	}
}

func shutdown(c io.Closer) {
	signalCh := make(chan os.Signal, 1)

	signal.Notify(signalCh,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGHUP,
	)

	<-signalCh
	log.Println("Shutting down...")

	c.Close()
}
