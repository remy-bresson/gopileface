package main

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type mockedDynamoDB struct {
	dynamodbiface.DynamoDBAPI
	ResponsePut dynamodb.PutItemOutput
	ResponseGet dynamodb.GetItemOutput
}

func (d mockedDynamoDB) PutItem(in *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return &d.ResponsePut, nil
}

func (d mockedDynamoDB) GetItem(in *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return &d.ResponseGet, nil
}

func TestHandler(t *testing.T) {
	t.Run("Successful Request", func(t *testing.T) {
		m := mockedDynamoDB{
			ResponsePut: dynamodb.PutItemOutput{},
			ResponseGet: dynamodb.GetItemOutput{},
		}

		d := dynamoInjection{
			ddb:   m,
			table: "test_table",
		}

		_, err := d.handler(events.APIGatewayProxyRequest{
			Resource:          "",
			Path:              "",
			HTTPMethod:        "",
			Headers:           map[string]string{},
			MultiValueHeaders: map[string][]string{},
			QueryStringParameters: map[string]string{
				"firstname": "toto",
				"lastname":  "titi",
			},
			MultiValueQueryStringParameters: map[string][]string{},
			PathParameters:                  map[string]string{},
			StageVariables:                  map[string]string{},
			RequestContext:                  events.APIGatewayProxyRequestContext{},
			Body:                            "",
			IsBase64Encoded:                 false,
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}
