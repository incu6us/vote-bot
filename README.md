vote-bot
---

Telegram bot for voting based on AWS serverless engines

### DynamoDB schema:
```
{
  Table: {
    AttributeDefinitions: [{
        AttributeName: "created_at",
        AttributeType: "N"
      },{
        AttributeName: "kind",
        AttributeType: "S"
      }],
    ItemCount: 0,
    KeySchema: [{
        AttributeName: "kind",
        KeyType: "HASH"
      },{
        AttributeName: "created_at",
        KeyType: "RANGE"
      }],
    TableName: "polls",
  }
}
```