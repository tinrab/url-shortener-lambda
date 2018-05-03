package main

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/teris-io/shortid"
)

const (
	LinksTableName = "UrlShortenerLinks"
	Region         = "us-east-1"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	ShortURL string `json:"short_url"`
}

type Link struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
  // Setup CORS header
  resp := events.APIGatewayProxyResponse{
    Headers: make(map[string]string),
  }
	resp.Headers["Access-Control-Allow-Origin"] = "*"
	// Parse request body
	rb := Request{}
	if err := json.Unmarshal([]byte(request.Body), &rb); err != nil {
		return resp, err
	}
	// Start DynamoDB session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(Region),
	})
	if err != nil {
		return resp, err
	}
	svc := dynamodb.New(sess)
	// Generate short url
	shortURL := shortid.MustGenerate()
	// Because "shorten" endpoint is reserved
	for shortURL == "shorten" {
		shortURL = shortid.MustGenerate()
	}
	link := &Link{
		ShortURL: shortURL,
		LongURL:  rb.URL,
	}
	// Marshal link to attribute value map
	av, err := dynamodbattribute.MarshalMap(link)
	if err != nil {
		return resp, err
	}
	// Put link
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(LinksTableName),
	}
	if _, err = svc.PutItem(input); err != nil {
		return resp, err
	}
	// Return short url
	response, err := json.Marshal(Response{ShortURL: shortURL})
	if err != nil {
		return resp, err
	}
  resp.StatusCode = http.StatusOK
  resp.Body = string(response)

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
