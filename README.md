vote-bot [![Build Status](https://travis-ci.org/incu6us/vote-bot.svg?branch=master)](https://travis-ci.org/incu6us/vote-bot)
---

Telegram bot for voting based on AWS DynamoDB


### Configuration
   Create configuration file `config.json` with the content:
    
```json
{
  "region": "eu-central-1",
  "dynamo": {
    "table": "polls"
  },
  "telegram": {
    "token": "bot-token",
    "bot_name":"bot-name",
    "user_ids": [
      {
        "some-user-name": 161500345
      }
    ]
  }
}
```

Description:
   * region - AWS region in which the dynamo's table shold be created
   * dynamo - setting for DynamoDB
   * telegram - Telegram settings
   * telegram.user_ids - users with IDs which will have an access to create a polls. User's key could be anything you want, but not th ID
   
   
### Create a poll
   To create poll use example below:
   ![Create poll](https://raw.githubusercontent.com/incu6us/vote-bot/master/doc/images/create_poll.png)
   
### Publish a poll
   To publish a poll you just need to type its name in group in which it is connected. After 4th typed symbol you'll find a popup with the poll.
   ![Publish poll](https://raw.githubusercontent.com/incu6us/vote-bot/master/doc/images/publish_poll.png)
   
   
   Result:
   
   ![Result](https://raw.githubusercontent.com/incu6us/vote-bot/master/doc/images/result.png)
    