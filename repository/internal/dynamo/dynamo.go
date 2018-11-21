package dynamo

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pkg/errors"
)

var (
	ErrBadPollName = errors.New("bad poll name")
)

type DB struct {
	tableName string
	client    *dynamodb.DynamoDB
}

func New(region, tableName string) (*DB, error) {
	awsCfg := aws.NewConfig().WithRegion(region).WithCredentials(credentials.NewEnvCredentials())

	sess, err := session.NewSession(awsCfg)
	if err != nil {
		return nil, errors.Wrap(err, "create session failed")
	}

	return &DB{tableName: tableName, client: dynamodb.New(sess)}, nil
}

func (db DB) CreateTable() error {
	_, err := db.client.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("subject"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("created_at"),
				AttributeType: aws.String("N"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("subject"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("created_at"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		TableName: aws.String(db.tableName),
	})

	return errors.Wrap(err, "create table failed")
}
func (db DB) DescribeTable() (string, error) {
	result, err := db.client.DescribeTable(&dynamodb.DescribeTableInput{TableName: aws.String(db.tableName)})
	if err != nil {
		return "", errors.Wrap(err, "failed to get table description")
	}

	return result.String(), nil
}

func (db DB) GetPolls() (*dynamodb.ScanOutput, error) {
	result, err := db.client.Scan(&dynamodb.ScanInput{
		TableName: aws.String(db.tableName),
		ScanFilter: map[string]*dynamodb.Condition{
			"subject": {
				ComparisonOperator: aws.String("NOT_NULL"),
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "get polls error")
	}

	return result, nil
}

func (db DB) GetPoll(subject string) (*dynamodb.QueryOutput, error) {
	if subject == "" {
		return nil, ErrBadPollName
	}

	result, err := db.client.Query(&dynamodb.QueryInput{
		TableName: aws.String(db.tableName),
		Limit:     aws.Int64(1),
		KeyConditions: map[string]*dynamodb.Condition{
			"subject": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{S: aws.String(subject)},
				},
			},
			"created_at": {
				ComparisonOperator: aws.String("GT"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{N: aws.String("0")},
				},
			},
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get poll with subject '%s' error", subject)
	}

	return result, nil
}

func (db DB) GetPollBeginsWith(subject string) (*dynamodb.ScanOutput, error) {
	if subject == "" {
		return nil, ErrBadPollName
	}

	result, err := db.client.Scan(&dynamodb.ScanInput{
		TableName: aws.String(db.tableName),
		ScanFilter: map[string]*dynamodb.Condition{
			"subject": {
				ComparisonOperator: aws.String("BEGINS_WITH"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{S: aws.String(subject)},
				},
			},
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get poll with subject '%s' error", subject)
	}

	return result, nil
}

func (db DB) GetPollByCreatedAt(createdAt int64) (*dynamodb.ScanOutput, error) {
	if createdAt == 0 {
		return nil, ErrBadPollName
	}

	result, err := db.client.Scan(&dynamodb.ScanInput{
		TableName: aws.String(db.tableName),
		ScanFilter: map[string]*dynamodb.Condition{
			"created_at": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{N: aws.String(strconv.FormatInt(createdAt, 10))},
				},
			},
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get poll with created_at field '%s' error", createdAt)
	}

	return result, nil
}

func (db DB) GetPollByOwner(subject, owner string) (*dynamodb.QueryOutput, error) {
	if subject == "" {
		return nil, ErrBadPollName
	}

	result, err := db.client.Query(&dynamodb.QueryInput{
		TableName: aws.String(db.tableName),
		Limit:     aws.Int64(1),
		KeyConditions: map[string]*dynamodb.Condition{
			"subject": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{S: aws.String(subject)},
				},
			},
			"created_at": {
				ComparisonOperator: aws.String("GT"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{N: aws.String("0")},
				},
			},
		},
		FilterExpression: aws.String("created_by = :o"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":o": {S: aws.String(owner)},
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get poll with subject '%s' error", subject)
	}

	return result, nil
}
func (db DB) CreatePoll(item map[string]*dynamodb.AttributeValue) error {
	_, err := db.client.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create items")
	}

	return nil
}

func (db DB) DeletePoll(subject string, createdAt int64) error {
	_, err := db.client.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"subject":    {S: aws.String(subject)},
			"created_at": {N: aws.String(strconv.FormatInt(createdAt, 10))},
		},
	})
	if err != nil {
		return errors.Wrapf(err, "failed to delete subject: %s", subject)
	}

	return nil
}

func (db DB) UpdateIsPublish(subject string, createdAt int64, isPublished bool) error {
	_, err := db.client.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"subject":    {S: aws.String(subject)},
			"created_at": {N: aws.String(strconv.FormatInt(createdAt, 10))},
		},
		UpdateExpression: aws.String("set is_published = :p"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":p": {BOOL: aws.Bool(isPublished)},
		},
	})

	return errors.Wrapf(err, "failed to update subject: %s", subject)
}

func (db DB) UpdateItems(subject string, createdAt int64, items []string) error {
	_, err := db.client.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"subject":    {S: aws.String(subject)},
			"created_at": {N: aws.String(strconv.FormatInt(createdAt, 10))},
		},
		UpdateExpression: aws.String("set #itemList = :i"),
		ExpressionAttributeNames: map[string]*string{
			"#itemList": aws.String("items"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":i": {SS: aws.StringSlice(items)},
		},
	})

	return errors.Wrapf(err, "failed to update subject: %s", subject)
}

func (db DB) UpdateVotes(subject string, createdAt int64, votes map[string]*dynamodb.AttributeValue) error {
	_, err := db.client.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"subject":    {S: aws.String(subject)},
			"created_at": {N: aws.String(strconv.FormatInt(createdAt, 10))},
		},
		UpdateExpression: aws.String("set votes = :v"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v": {M: votes},
		},
	})

	return errors.Wrapf(err, "failed to update votest: %s", subject)
}
