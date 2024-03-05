package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

func init() {
	ginLambda = ginadapter.New(setupRouter())
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, request)
}

func setupRouter() *gin.Engine {
	// stdout and stderr are sent to AWS CloudWatch Logs
	log.Printf("Gin cold start")
	r := gin.Default()
	r.GET("/user/:id", HandleUser)
	r.GET("/user", HandleAll)
	r.GET("/count", HandleCount)
	r.POST("/", HandleInsert)

	// Default home page
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Home page!",
		})
	})
	return r
}

func main() {
	// Starts the handler for AWS Lambda
	lambda.Start(Handler)
}
