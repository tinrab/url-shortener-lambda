package main

import (
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const (
	LinksTableName = "UrlShortenerLinks"
	Region         = "us-east-1"
)

var (
	ErrMissingShortUrl error = errors.New("short_url path parameter is missing")
)

type Link struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get short_url parameter
	shortURL, ok := request.PathParameters["short_url"]
	if !ok {
		return events.APIGatewayProxyResponse{}, ErrMissingShortUrl
	}
	// Start DynamoDB session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(Region),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	svc := dynamodb.New(sess)
	// Read link
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(LinksTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"short_url": {
				S: aws.String(shortURL),
			},
		},
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	// Unmarshal link item
	link := Link{}
	if err := dynamodbattribute.UnmarshalMap(result.Item, &link); err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	// Redirect to long URL
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusPermanentRedirect,
		Headers: map[string]string{
			"location": link.LongURL,
		},
	}, nil
}

func main() {
	lambda.Start(Handler)
}
