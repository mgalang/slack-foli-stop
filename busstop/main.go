package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

// RecordItem inferface.
type RecordItem struct {
	Destinationdisplay    string
	Aimedarrivaltime      int64
	Expecteddeparturetime int64
}

// SiriJSON interface.
type SiriJSON struct {
	Status     string
	Servertime int64
	Result     []RecordItem
}

// Get JSON from URL helper method.
func getJSON(path string, result interface{}) error {
	resp, err := http.Get(path)

	if resp.StatusCode != 200 {
		panic("Server error")
	}

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(result)
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context) (Response, error) {
	var buf bytes.Buffer

	j := SiriJSON{}

	// TODO, use parameter.
	var stopcode = "t4"

	err := getJSON("https://data.foli.fi/siri/sm/"+stopcode, &j)

	if err != nil {
		return Response{StatusCode: 404}, err
	}

	var count = len(j.Result)

	if count > 3 {
		count = 3
	}

	var responseText = ""

	for index := 0; index < count; index++ {
		t := time.Unix(j.Result[index].Expecteddeparturetime, 0)

		responseText += "Destination: " + j.Result[index].Destinationdisplay + ", Leaving at: " + t.Format("01:01:01")
	}

	body, err := json.Marshal(map[string]interface{}{
		"text": responseText,
	})

	if err != nil {
		return Response{StatusCode: 404}, err
	}

	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
