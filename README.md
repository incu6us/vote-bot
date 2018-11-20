vote-bot
---

Telegram bot for voting based on AWS DynamoDB


### Configuration
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
    