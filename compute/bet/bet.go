package main

import (
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/remy-bresson/gopileface/commons"
	"github.com/remy-bresson/gopileface/maputils"
	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type dynamoInjection struct {
	ddb   dynamodbiface.DynamoDBAPI
	table string
}

func checkBet(bet string) (string, error) {
	if bet != "pile" && bet != "face" {
		return "", errors.New("vous devez jouer pile ou face")
	}
	return bet, nil
}

func (d *dynamoInjection) checkAndUpdateAmount(uid string, amount int) (int, error) {
	result, err := d.ddb.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(d.table),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(uid),
			},
		},
	})

	if err != nil {
		log.Error("Got error calling GetItem: %s", err)
		return 0, errors.New("error during user retrieving phase")
	}

	if result.Item == nil {
		log.Error("User not found")
		return 0, errors.New("unable to find user")
	}

	person := commons.Person{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &person)
	if err != nil {
		log.Error(err)
		return 0, errors.New("unable to unmarshall user information")
	}

	if person.Amount < amount {
		log.Error("No more money")
		return 0, errors.New("no more money")
	}

	newAmount := person.Amount - amount

	updateItem := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":amount": {
				N: aws.String(strconv.Itoa(newAmount)),
			},
		},
		TableName: aws.String(d.table),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(uid),
			},
		},
		ReturnValues:     aws.String("NONE"),
		UpdateExpression: aws.String("set money = :amount"),
	}

	_, err = d.ddb.UpdateItem(updateItem)
	if err != nil {
		log.Error("Error during amount update in dynamo db", err)
		return 0, errors.New("error during amount update in dynamo db")
	}

	return newAmount, nil
}

func (d *dynamoInjection) handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	/* ------------------------------------------------------------------ */
	/* Ectracting lastname + firstname from query                         */
	uid, err := maputils.ExtractSomethingFromMap(request.Headers, "uid", true)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       string("you must pass an uid"),
		}, nil
	}
	log.Info("======>uid :", uid)

	amount, err := maputils.ExtractSomethingFromMap(request.QueryStringParameters, "amount", true)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       string("you must pass an amount"),
		}, nil
	}
	amountInteger, _ := strconv.Atoi(amount)
	log.Info("======>amount :", amountInteger)

	bet, err := maputils.ExtractSomethingFromMap(request.QueryStringParameters, "bet", true)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       string("you must pass an bet"),
		}, nil
	}
	log.Info("======>bet :", bet)
	/* ------------------------------------------------------------------ */

	/* ------------------------------------------------------------------ */
	/* Check if bet value is valid one                                    */
	_, err = checkBet(bet)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       string("you must play pile or face, nothing else"),
		}, nil
	}
	/* ------------------------------------------------------------------ */

	/* ------------------------------------------------------------------ */
	/* Check if user has still credit                                     */
	currentAmount, err := d.checkAndUpdateAmount(uid, amountInteger)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       string("no more money"),
		}, nil
	}
	/* ------------------------------------------------------------------ */

	/* ------------------------------------------------------------------ */
	/* Get a random float between 0 and 1                                 */
	tirage := rand.Intn(10)

	var resultatTirage string

	if tirage > 5 {
		/* Player win */
		resultatTirage = bet
		gain := amountInteger * 2
		currentAmount = currentAmount + gain

		updateItem := &dynamodb.UpdateItemInput{
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":amount": {
					N: aws.String(strconv.Itoa(currentAmount)),
				},
			},
			TableName: aws.String(d.table),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(uid),
				},
			},
			ReturnValues:     aws.String("NONE"),
			UpdateExpression: aws.String("set money = :amount"),
		}

		_, err = d.ddb.UpdateItem(updateItem)
		if err != nil {
			log.Error("Error during amount update after win in dynamo db")
			return events.APIGatewayProxyResponse{
				StatusCode: 501,
				Body:       string("Error during amount update after win in dynamo db"),
			}, nil
		}

	} else {
		if bet == "pile" {
			resultatTirage = "face"
		} else {
			resultatTirage = "pile"
		}
		log.Info("Player has loose")
	}
	/* ------------------------------------------------------------------ */

	/* ------------------------------------------------------------------ */
	/* Build output struc                                                 */
	var output map[string]string = make(map[string]string)
	output["result"] = resultatTirage
	output["amount"] = strconv.Itoa(currentAmount)
	ret, _ := json.Marshal(output)
	/* ------------------------------------------------------------------ */

	/* ------------------------------------------------------------------ */
	/* Manage return                                                      */
	return events.APIGatewayProxyResponse{
		Body:       string(ret),
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
