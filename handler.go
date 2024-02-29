package main

import (
	"errors"
	"fmt"
	"net/http"

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

type databasePut struct {
	User_id  string `json:"user_id"`
	Eligible string `json:"eligible"`
}

type DatabasePutSlice []databasePut

const tableName string = "user_data"

func HandleUser(c *gin.Context) {

	// Create a new instance of databaseInfo
	var info databaseInfo
	info.User_id = c.Param("id")

	if err := info.dynamoGet("/user/"); err != nil {
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

	if err := info.dynamoGet("/count"); err != nil {
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

func HandleInsert(c *gin.Context) {
	// Create a new instance of DatabasePutSlice
	var infoSlice DatabasePutSlice

	// Bind the JSON request body to the info variable
	if err := c.BindJSON(&infoSlice); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
		return
	}

	// Call the dynamoPost method to insert the data into the database
	if err := infoSlice.dynamoPost(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond with success message
	c.JSON(http.StatusOK, gin.H{"message": "Data inserted successfully"})
}

func (infoSlice *DatabasePutSlice) dynamoPost() error {
	// Start DynamoDb connection=
	sess := session.Must(session.NewSession())

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	for _, info := range *infoSlice {

		av, err := dynamodbattribute.MarshalMap(info)
		if err != nil {
			return fmt.Errorf("failed to marshal map: %v", err)
		}
		// Create new item
		_, err = svc.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      av,
		})

		if err != nil {
			return fmt.Errorf("failed to insert item: %v", av)
		}
	}
	return nil
}

func (info *databaseInfo) dynamoGet(relativePath string) error {

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
