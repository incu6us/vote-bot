package dynamo

import (
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	pollKind = "poll"
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

func (db DB) DescribeTable() (string, error) {
	result, err := db.client.DescribeTable(&dynamodb.DescribeTableInput{TableName: aws.String(db.tableName)})
	if err != nil {
		return "", errors.Wrap(err, "feiled to get table description")
	}

	return result.String(), nil
}

func (db DB) GetPolls(limitRows int64) (*dynamodb.QueryOutput, error) {
	if limitRows == 0 {
		return nil, ErrBadLimit
	}

	result, err := db.client.Query(&dynamodb.QueryInput{
		TableName:        aws.String(db.tableName),
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int64(limitRows),
		KeyConditions: map[string]*dynamodb.Condition{
			"kind": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(pollKind),
					},
				},
			},
			"created_at": {
				ComparisonOperator: aws.String("LE"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						N: aws.String(strconv.FormatInt(time.Now().UnixNano(), 10)),
					},
				},
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "get polls error")
	}

	return result, nil
}

func (db DB) GetPoll(pollName string) (*dynamodb.GetItemOutput, error) {
	if pollName == "" {
		return nil, ErrBadPollName
	}

	result, err := db.client.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"subject": {
				S: aws.String(pollName),
			},
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get poll with subject: %s error", pollName)
	}

	return result, nil
}
