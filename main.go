package main

import (
	"log"
	cfg "vote-bot/config"
	"vote-bot/repository"
)

const (
	cfgPrefix = "VB"
	cfgFile   = "config.json"
)

const (
	indexName = "subject"
	pageLimit = 5
)

type Answer struct {
	Item  string   `json:"item"`
	Users []string `json:"users"`
}

type Vote struct {
	Subject string    `json:"subject"`
	Items   []string  `json:"items"`
	Answers []*Answer `json:"answers"`
}

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

	repo, err := repository.New(region, tableName)
	if err != nil {
		log.Printf("failed to initiate repository: %s", err)
		return
	}

	polls, err := repo.GetPolls()
	if err != nil {
		log.Printf("can't get polls: %s", err)
		return
	}

	for _, poll := range polls {
		log.Printf("data: %+v", *poll)
	}

	resp, err := repo.DescribeTable()
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("description:", resp)
}
