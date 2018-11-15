package dynamo

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/pkg/errors"
)

var (
	ErrBadPollName = errors.New("bad poll name")
	ErrBadLimit    = errors.New("get poll limit with 0 row can't be done")
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
		return "", errors.Wrap(err, "feiled to get table description")
	}

	return result.String(), nil
}

func (db DB) GetPolls(limitRows int64) (*dynamodb.ScanOutput, error) {
	if limitRows == 0 {
		return nil, ErrBadLimit
	}

	result, err := db.client.Scan(&dynamodb.ScanInput{
		TableName: aws.String(db.tableName),
		Limit:     aws.Int64(limitRows),
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

func (db DB) GetPoll(pollName string) (*dynamodb.QueryOutput, error) {
	if pollName == "" {
		return nil, ErrBadPollName
	}

	result, err := db.client.Query(&dynamodb.QueryInput{
		TableName: aws.String(db.tableName),
		Limit:     aws.Int64(1),
		KeyConditions: map[string]*dynamodb.Condition{
			"subject": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(pollName),
					},
				},
			},
			"created_at": {
				ComparisonOperator: aws.String("GT"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						N: aws.String("0"),
					},
				},
			},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":s": {
				S: aws.String(pollName),
			},
		},
		FilterExpression: aws.String("subject = :s"),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get poll with subject '%s' error", pollName)
	}

	return result, nil
}

func (db DB) CreatePoll(poll interface{}) error {
	fmt.Printf("!!!: %+v", poll)
	item, err := dynamodbattribute.MarshalMap(poll)
	if err != nil {
		return errors.Wrap(err, "create poll error")
	}

	_, err = db.client.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})

	return err
}
