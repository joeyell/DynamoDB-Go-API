package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gin-gonic/gin"
)

type databaseInfo struct {
	User_id  string `json:"user_id"`
	Eligible bool   `json:"eligible"`
	Count    int64  `json:"count"`
}

const tableName string = "user_data"

func HandleUser(c *gin.Context) {
	log.Println("Received a GET request to /user/:id")

	var info databaseInfo
	info.User_id = c.Param("id")

	relativePath := c.Request.URL.Path

	info.dynamoStart(relativePath)

	c.JSON(200, gin.H{
		"message": info.Eligible,
	})
}

func HandleCount(c *gin.Context) {
	// Log that a GET request to /count has been received
	log.Println("Received a GET request to /count")

	// Create a new instance of databaseInfo
	var info databaseInfo

	// Extract the relative path from the request URL
	relativePath := c.Request.URL.Path

	// Call the dynamoStart method of the databaseInfo instance
	// to perform some action related to DynamoDB
	if err := info.dynamoStart(relativePath); err != nil {
		// If there's an error, return an internal server error
		c.JSON(500, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	// Return a JSON response with a message indicating whether
	// the request is eligible or not
	c.JSON(200, gin.H{
		"message": info.Eligible,
	})
}

func (info *databaseInfo) dynamoStart(relativePath string) error {

	// Check if the relativePath is "/count"
	if relativePath != "/count" {
		// If it's not "/count", return an error
		return errors.New("invalid endpoint")
	}
	switch relativePath {
	case "/user/:id":
		info.dynamoUser()
	case "/count":
		info.dynamoCount()

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

	// Convert ScanOuput to int64
	count := *result.Count

	info.Count = count

	if err != nil {
		return fmt.Errorf("failed to call GetItem: %s", err)
	}

	if count == 0 {
		return fmt.Errorf("database contains no values: %s", err)
	}

	return nil
}
