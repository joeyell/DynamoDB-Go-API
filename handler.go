package main

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gin-gonic/gin"
)

type databaseInfo struct {
	User_id  string `json:"user_id"`
	Eligible string `json:"eligible"`
	Count    int64  `json:"count"`
}

const tableName string = "user_data"

func HandleUser(c *gin.Context) {

	// Create a new instance of databaseInfo
	var info databaseInfo
	info.User_id = c.Param("id")

	if err := info.dynamoStart("/user/"); err != nil {
		// If there's an error, return an internal server error
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": info.Eligible,
	})
}

func HandleCount(c *gin.Context) {

	// Create a new instance of databaseInfo
	var info databaseInfo

	if err := info.dynamoStart("/count"); err != nil {
		// If there's an error, return an internal server error
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": info.Count,
	})
}

func (info *databaseInfo) dynamoStart(relativePath string) error {

	switch relativePath {
	case "/user/":
		info.dynamoUser()
	case "/count":
		info.dynamoCount()
	default:
		return errors.New("invalid endpoint")
	}
	return nil
}

func (info *databaseInfo) dynamoUser() error {

	// Start DynamoDB connection
	sess := session.Must(session.NewSession())

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Retrieve result
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {
				S: aws.String(info.User_id),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to call GetItem: %s", err)
	}

	if len(result.Item) == 0 {
		return fmt.Errorf("item not found for user_id: %s", info.User_id)
	}

	if err := dynamodbattribute.UnmarshalMap(result.Item, &info); err != nil {
		return fmt.Errorf("failed to unmarshal Record: %s", err)
	}
	return nil
}

func (info *databaseInfo) dynamoCount() error {

	// Start DynamoDB connection
	sess := session.Must(session.NewSession())

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Retrieve result
	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String(tableName),
		Select:    aws.String("COUNT"),
	})

	// Handle errors
	if err != nil {
		return fmt.Errorf("failed to call Scan: %s", err)
	}

	// Convert ScanOuput to int64
	count := *result.Count

	// Assign count to info.Count
	info.Count = count

	// Handle empty database
	if count == 0 {
		return fmt.Errorf("database contains no values: %s", err)
	}
	return nil
}
