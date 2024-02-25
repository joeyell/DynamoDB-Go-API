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
	"github.com/valyala/fastjson"
)

type UserData struct {
	User_id  string `json:"user_id"`
	Eligible bool   `json:"eligible"`
}

const tableName string = "user_data"

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ApiResponse := events.APIGatewayProxyResponse{}
	// Switch for identifying the HTTP request
	switch request.HTTPMethod {
	case "GET":
		// Obtain the QueryStringParameter
		name := string(request.QueryStringParameters["name"])
		if name != "" {
			user := UserData{User_id: name}
			err := user.dynamoGET()
			if err != nil {

				ApiResponse = events.APIGatewayProxyResponse{
					StatusCode: 500,
					Body:       fmt.Sprintf("Error: %s", err.Error()),
				}
			} else {
				jsonString, _ := json.Marshal(user)
				ApiResponse = events.APIGatewayProxyResponse{
					StatusCode: 200,
					Body:       string(jsonString),
				}
			}
		} else {
			ApiResponse = events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       "Error: Query Parameter name missing",
			}
		}

	case "POST":
		//validates json and returns error if not working
		err := fastjson.Validate(request.Body)

		if err != nil {
			body := "Error: Invalid JSON payload ||| " + fmt.Sprint(err) + " Body Obtained" + "||||" + request.Body
			ApiResponse = events.APIGatewayProxyResponse{Body: body, StatusCode: 500}
		} else {
			ApiResponse = events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}
		}

	}
	// Response
	return ApiResponse, nil
}

func (user *UserData) dynamoGET() error {
	sess := session.Must(session.NewSession())

	// Defer the closure of the session
	//defer sess.Close()

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
