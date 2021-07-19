package main

import (
	"errors"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/remy-bresson/gopileface/maputils"
	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type dynamoInjection struct {
	ddb   dynamodbiface.DynamoDBAPI
	table string
}

type Person struct {
	ID        string `dynamodbav:"id"`
	Firstname string `dynamodbav:"firstanme"`
	Latname   string `dynamodbav:"lastname"`
	Amount    int    `dynamodbav:"money"`
}

func checkBet(bet string) (string, error) {
	if bet != "pile" && bet != "face" {
		return "", errors.New("vous devez jouer pile ou face")
	}
	return bet, nil
}

func (d *dynamoInjection) handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	/* ------------------------------------------------------------------ */
	/* Ectracting lastname + firstname from query                         */
	uid, err := maputils.ExtractSomethingFromMap(request.Headers, "uid", true)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	log.Info("======>uid :", uid)

	amount, err := maputils.ExtractSomethingFromMap(request.QueryStringParameters, "amount", true)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	log.Info("======>amount :", amount)

	bet, err := maputils.ExtractSomethingFromMap(request.QueryStringParameters, "bet", true)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	log.Info("======>bet :", bet)
	/* ------------------------------------------------------------------ */

	/* ------------------------------------------------------------------ */
	/* Check if bet value is valid one                                    */
	_, err = checkBet(bet)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	/* ------------------------------------------------------------------ */

	/* ------------------------------------------------------------------ */
	/* Manage return                                                      */
	return events.APIGatewayProxyResponse{
		Body:       string("OK"),
		StatusCode: 200,
	}, nil
	/* ------------------------------------------------------------------ */
}

func main() {
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSession())

	// Create DynamoDB client
	ddbClient := dynamodb.New(sess)

	d := dynamoInjection{
		ddb:   ddbClient,
		table: os.Getenv("TableName"),
	}

	lambda.Start(d.handler)
}
