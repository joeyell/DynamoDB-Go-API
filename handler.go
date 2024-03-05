package main

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gin-gonic/gin"
)

type databaseInfo struct {
	Count int64 `json:"count"`
}

type User struct {
	User_id string   `json:"user_id"`
	Data    UserData `json:"data"`
}

type DynamoDBItem struct {
	UserID     string                 `json:"user_id"`
	Attributes map[string]interface{} `json:"data"`
}

type UserData struct {
	EntryDate string `json:"entry_date"`
	Food      string `json:"food"`
}

type DatabasePutSlice []User

type DynamoDBItemSlice []DynamoDBItem

const tableName string = "user_data"

func HandleUser(c *gin.Context) {

	// Create a new instance of databaseInfo
	var info User
	info.User_id = c.Param("id")

	if err := info.getUser(); err != nil {
		// If there's an error, return an internal server error
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": info,
	})
}

func HandleAll(c *gin.Context) {

	var info DynamoDBItemSlice

	if err := info.getAll(); err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
	}
	c.JSON(200, gin.H{
		"message": info,
	})
}

func HandleCount(c *gin.Context) {

	// Create a new instance of databaseInfo
	var info databaseInfo

	if err := info.getCount(); err != nil {
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

func (info *User) getUser() error {

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

	// Check for errors returned by GetItem
	if err != nil {
		return fmt.Errorf("failed to call GetItem: %s", err)
	}

	// Check if the item was found
	if result.Item == nil {
		return fmt.Errorf("no item found for user_id: %s", info.User_id)
	}
	// Check if "data" attribute exists in the retrieved item
	dataAttr := result.Item["data"]

	// Dereference the pointer to string if it's not nil
	info.Data.Food = *dataAttr.M["food"].S
	info.Data.EntryDate = *dataAttr.M["entry_date"].S

	return nil
}

func (info *databaseInfo) getCount() error {

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

func (info *DynamoDBItemSlice) getAll() error {

	// Start DynamoDB connection
	sess := session.Must(session.NewSession())

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})

	if err != nil {
		return fmt.Errorf("error retrieving all data: %s", err)
	}

	for _, item := range result.Items {
		// Unmarshal the item into a DynamoDBItem struct
		var dynamoDBItem DynamoDBItem
		dynamodbattribute.UnmarshalMap(item, &dynamoDBItem.Attributes)

		// Append the DynamoDBItem to the items slice
		*info = append(*info, dynamoDBItem)
	}

	return nil
}
