package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

var (
	// DefaultHTTPGetAddress Default Address
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// ErrNoIP No IP found in response
	ErrNoIP = errors.New("No IP in HTTP response")

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("Non 200 Response found")

	tableName = "room_conditions"
)

type RoomConditions struct {
	DeviceNames          string  `json:"device_names"`
	CreatedAt            string  `json:"created_at"`
	Humid                int     `json:"humid"`
	HumidCreatedAt       string  `json:"humid_created_at"`
	Illuminance          float64 `json:"illuminance"`
	IlluminanceCreatedAt string  `json:"illuminance_created_at"`
	Temperature          float64 `json:"temperature"`
	TemperatureCreatedAt string  `json:"temperature_created_at"`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	data()
	fmt.Println(request.QueryStringParameters)
	resp, err := http.Get(DefaultHTTPGetAddress)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	if resp.StatusCode != 200 {
		return events.APIGatewayProxyResponse{}, ErrNon200Response
	}

	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	if len(ip) == 0 {
		return events.APIGatewayProxyResponse{}, ErrNoIP
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Hello, %v", string(ip)),
		StatusCode: 200,
	}, nil
}

func data() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	deviceName := "living-room-remo"
	// createdAt := "2019-05-28T03:42:22+09:00"
	var minTemperature float64 = 24.0
	// a1 := "2019-05-26T00:00:00+09:00"
	// a2 := "2019-05-28T00:00:00+09:00"
	date := "2019-05-26"

	filt := expression.Name("device_names").Equal(expression.Value(deviceName))
	// filt2 := expression.Name("created_at").Equal(expression.Value(createdAt))
	// filt2 := expression.Name("created_at").Contains("2019-05-28")
	filt2 := expression.Name("created_at").Contains(date)

	proj := expression.NamesList(expression.Name("device_names"), expression.Name("created_at"), expression.Name("temperature"), expression.Name("humid"))

	expr, err := expression.NewBuilder().WithFilter(filt).WithFilter(filt2).WithProjection(proj).Build()
	if err != nil {
		fmt.Println("Got error building expression:")
		fmt.Println(err.Error())
	}

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
	}

	// Make the DynamoDB Query API call
	result, err := svc.Scan(params)
	if err != nil {
		fmt.Println("Query API call failed:")
		fmt.Println((err.Error()))
		os.Exit(1)
	}
	// snippet-end:[dynamodb.go.scan_items.call]

	// snippet-start:[dynamodb.go.scan_items.process]
	numItems := 0

	for _, i := range result.Items {
		item := RoomConditions{}

		err = dynamodbattribute.UnmarshalMap(i, &item)

		if err != nil {
			fmt.Println("Got error unmarshalling:")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// Which ones had a higher rating than minimum?
		if item.Temperature > minTemperature {
			// Or it we had filtered by rating previously:
			//   if item.Year == year {
			numItems++

			fmt.Println("Device: ", item.DeviceNames)
			fmt.Println("CreatedAt:", item.CreatedAt)
			fmt.Println("humid:", item.Humid)
			fmt.Println("Temperature: ", item.Temperature)
			fmt.Println()
		}
	}

	fmt.Println("Found", numItems)
}

func main() {
	lambda.Start(handler)
}
