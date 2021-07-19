package main

import (
	"os"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/remy-bresson/gopileface/commons"
	"github.com/remy-bresson/gopileface/maputils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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

func (d *dynamoInjection) getUniqID() string {
	res := uuid.NewV4().String()

	for {
		result, err := d.ddb.GetItem(&dynamodb.GetItemInput{
			TableName: aws.String(d.table),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(res),
				},
			},
		})

		if err != nil {
			log.Fatalf("Got error calling GetItem: %s", err)
		}
		if result.Item != nil {
			// Id is already in used
			log.Warn("Id already in used : ", res)
			res = uuid.NewV4().String()
		} else {
			// Find an empty ID, return it!
			return res
		}
	}
}

func (d *dynamoInjection) handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	/* ------------------------------------------------------------------ */
	/* Ectracting lastname + firstname from query                         */
	lastname, err := maputils.ExtractSomethingFromMap(request.QueryStringParameters, "lastname", true)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	log.Info("======>lastname :", lastname)

	firstname, err := maputils.ExtractSomethingFromMap(request.QueryStringParameters, "firstname", true)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	log.Info("======>firstname :", firstname)
	/* ------------------------------------------------------------------ */

	/* ------------------------------------------------------------------ */
	/* Create Person struc withj incoming information + uuid              */
	var o commons.Person
	o.ID = d.getUniqID()
	o.Firstname = firstname
	o.Latname = lastname
	o.Amount = 10
	/* ------------------------------------------------------------------ */

	/* ------------------------------------------------------------------ */
	/* Create dynamodb input                                              */
	item, err := dynamodbattribute.MarshalMap(o)
	if err != nil {
		log.Error("Could not marshall dynamodb input : \n", o)
		return events.APIGatewayProxyResponse{}, err
	}

	input := dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(d.table),
	}
	/* ------------------------------------------------------------------ */

	/* ------------------------------------------------------------------ */
	/* Add Person into dynamodb                                           */
	_, err = d.ddb.PutItem(&input)

	if err != nil {
		log.Error("Error when writing on dynamodb")
		return events.APIGatewayProxyResponse{}, err
	}
	/* ------------------------------------------------------------------ */

	// body, err := json.Marshal(res.Attributes)

	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	/* ------------------------------------------------------------------ */
	/* Manage return                                                      */
	return events.APIGatewayProxyResponse{
		Body:       string(o.ID),
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
