package repository

import (
	"testing"
)

func TestRepository(t *testing.T) {
	// awsCfg := aws.NewConfig().WithRegion(region).WithCredentials(credentials.NewEnvCredentials())
	//
	// sess, err := session.NewSession(awsCfg)
	// if err != nil {
	// 	log.Printf("create session failed: %s", err)
	// 	return
	// }
	//
	// ddb := dynamodb.New(sess)

	// result, err := ddb.ListTables(&dynamodb.ListTablesInput{})
	// if err != nil {
	// 	log.Printf("get list error: %s", err)
	// }
	//
	// log.Printf("result:%+v", result)

	// result1, err := ddb.DescribeTable(&dynamodb.DescribeTableInput{TableName: aws.String(tableName)})
	// if err != nil {
	// 	log.Printf("ddb error: %s", err)
	// }
	//
	// log.Printf("result: %+v", result1)
	//
	// result, err := ddb.GetItem(&dynamodb.GetItemInput{
	// 	TableName: aws.String(tableName),
	// 	Key: map[string]*dynamodb.AttributeValue{
	// 		"subject": {
	// 			S: aws.String(""),
	// 		},
	// 	},
	// })
	// if err != nil {
	// 	log.Printf("get data error: %s", err)
	// }
	//
	// log.Printf("data: %+v", result)

	// result, err := ddb.Query(&dynamodb.QueryInput{
	// 	TableName:        aws.String(tableName),
	// 	ScanIndexForward: aws.Bool(false),
	// 	Limit:            aws.Int64(10),
	// 	KeyConditions: map[string]*dynamodb.Condition{
	// 		"kind": {
	// 			ComparisonOperator: aws.String("EQ"),
	// 			AttributeValueList: []*dynamodb.AttributeValue{
	// 				{
	// 					S: aws.String("vote"),
	// 				},
	// 			},
	// 		},
	// 		"created_at": {
	// 			ComparisonOperator: aws.String("LE"),
	// 			AttributeValueList: []*dynamodb.AttributeValue{
	// 				{
	// 					N: aws.String(strconv.FormatInt(time.Now().UnixNano(), 10)),
	// 				},
	// 			},
	// 		},
	// 	},
	// })
	// if err != nil {
	// 	log.Printf("query database error: %s", err)
	// 	return
	// }

	// result, err := ddb.Scan(&dynamodb.ScanInput{
	// 	TableName: aws.String(tableName),
	// 	Limit:     aws.Int64(pageLimit),
	// })
	// if err != nil {
	// 	log.Printf("list tables error: %s", err)
	// 	return
	// }
	// log.Printf("DATA: %+v", result)

	// repo, err := repository.New(region, tableName)
	// if err != nil {
	// 	log.Printf("failed to initiate repository: %s", err)
	// 	return
	// }
	//
	// _, err = repo.DescribeTable()
	// if err != nil {
	// 	if awsErr, ok := errors.Cause(err).(awserr.Error); ok {
	// 		switch awsErr.Code() {
	// 		case dynamodb.ErrCodeResourceNotFoundException:
	// 			if err1 := repo.CreateTable(); err != nil {
	// 				log.Printf("create table error: %s", err1)
	// 			}
	// 			log.Println("table created")
	// 		}
	// 	} else {
	// 		log.Println(err)
	// 		return
	// 	}
	// }

	// desc, _ := repo.DescribeTable()
	//
	// log.Printf("desc: %+v", desc)

	// polls, err := repo.GetPolls()
	// if err != nil {
	// 	log.Printf("can't get polls: %s", err)
	// 	return
	// }
	// //
	// for _, poll := range polls {
	// 	log.Printf("data: %+v", *poll)
	// }

	// err = repo.CreatePoll("test", "me", []string{"1", "2"})
	// if err != nil {
	// 	log.Printf("poll created failed: %s", err)
	// 	return
	// }
	// log.Println("poll created")

	// err = repo.DeletePoll("test", "me")
	// if err != nil {
	// 	log.Printf("poll delete failed: %s", err)
	// 	return
	// }
	// log.Println("poll deleted")

	// err = repo.UpdatePollIsPublished("test1", "me", true)
	// if err != nil {
	// 	log.Printf("poll update failed: %s", err)
	// 	return
	// }
	// log.Println("poll updated")

	// err = repo.UpdatePollItems("test1", "me", []string{"qw", "rt"})
	// if err != nil {
	// 	log.Printf("poll update failed: %s", err)
	// 	return
	// }
	// log.Println("poll updated")
}
