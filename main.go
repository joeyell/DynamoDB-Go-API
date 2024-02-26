package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	//"github.com/valyala/fastjson"
)

type UserData struct {
	User_id  string `json:"user_id"`
	Eligible bool   `json:"eligible"`
}

const tableName string = "user_data"

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Switch for identifying the HTTP request
	switch request.HTTPMethod {
	case "GET":
		return HandleGetRequest(request)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Error: Query Parameter name missing",
		}, nil
	}
}

func HandleGetRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// Obtain query string
	name := request.QueryStringParameters["name"]
	if name == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Error: Query Parameter name missing",
		}, nil
	}

	// Retrieve user data
	userData, err := getUserData(name)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("Error: %s", err.Error()),
		}, nil
	}

	// Format response
	jsonString, err := formatResponse(userData)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("Error formatting response: %s", err.Error()),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       jsonString,
	}, nil
}

func getUserData(name string) (UserData, error) {
	user := UserData{User_id: name}
	err := user.dynamoGet()
	return user, err
}

func formatResponse(userData UserData) (string, error) {
	jsonString, err := json.Marshal(userData)
	if err != nil {
		return "", err
	}
	return string(jsonString), nil
}

func (user *UserData) dynamoGet() error {

	// Start DynamoDB connection
	sess := session.Must(session.NewSession())

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {
				S: aws.String(user.User_id),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to call GetItem: %s", err)
	}

	if len(result.Item) == 0 {
		return fmt.Errorf("item not found for user_id: %s", user.User_id)
	}

	if err := dynamodbattribute.UnmarshalMap(result.Item, &user); err != nil {
		return fmt.Errorf("failed to unmarshal Record: %s", err)
	}
	return nil
}

func main() {
	// Starts the handler for AWS Lambda
	lambda.Start(HandleRequest)
}
