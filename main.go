package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	cf "github.com/aws/aws-sdk-go/service/cloudformation"
)

// Marshal creates a binary payload from a WavefrontEvent.
// The method returns either the binary payload or an error.
func (r *WavefrontEvent) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// WavefrontEvent represents the data structure expected by Wavefront
type WavefrontEvent struct {
	Table       string      `json:"table"`
	Name        string      `json:"name"`
	Annotations Annotations `json:"annotations"`
	StartTime   int64       `json:"startTime"`
	EndTime     int64       `json:"endTime"`
}

// Annotations are the annotations needed by Wavefront
type Annotations struct {
	Severity string `json:"severity"`
	Type     string `json:"type"`
	Details  string `json:"details"`
}

var (
	wavefrontEventURL = os.Getenv("WAVEFRONT_URL")
	wavefrontAPIToken = os.Getenv("WAVEFRONT_TOKEN")
)

func handler(request events.SNSEvent) error {
	for _, i := range request.Records {
		stackEvent, err := parseCFMessage(i.SNS.Message)
		if err != nil {
			return err
		}
		err = createWavefrontEvent(stackEvent)
		if err != nil {
			return err
		}
	}
	return nil
}

func createWavefrontEvent(event cf.StackEvent) error {
	evt := WavefrontEvent{
		Table:     "",
		Name:      fmt.Sprintf("CloudFormation Event for %s", *event.StackName),
		StartTime: event.Timestamp.Unix(),
		EndTime:   event.Timestamp.Unix() + 1,
		Annotations: Annotations{
			Severity: "info",
			Type:     "CloudFormation",
			Details:  fmt.Sprintf("Event ID %s (%s)", *event.EventId, *event.ResourceStatus),
		},
	}

	payload, err := evt.Marshal()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", wavefrontEventURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", wavefrontAPIToken))
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("HTTP response %d (%s)", res.StatusCode, res.Status)
	}

	return nil
}

func parseCFMessage(message string) (cf.StackEvent, error) {
	elements := strings.Split(message, "\n")
	elementMap := make(map[string]*string)

	for _, element := range elements {
		if len(element) > 0 && strings.Contains(element, "=") {
			items := strings.Split(element, "=")
			ptrString := strings.ReplaceAll(items[1], "'", "")
			elementMap[items[0]] = &ptrString
		}
	}

	if len(elementMap) == 0 {
		return cf.StackEvent{}, fmt.Errorf("stack message contains no fields")
	}

	time, err := parseDate(elementMap["Timestamp"])
	if err != nil {
		return cf.StackEvent{}, err
	}

	return cf.StackEvent{
		StackId:              elementMap["StackId"],
		EventId:              elementMap["EventId"],
		LogicalResourceId:    elementMap["LogicalResourceId"],
		PhysicalResourceId:   elementMap["PhysicalResourceId"],
		ResourceProperties:   elementMap["ResourceProperties"],
		ResourceStatus:       elementMap["ResourceStatus"],
		ResourceStatusReason: elementMap["ResourceStatusReason"],
		ResourceType:         elementMap["ResourceType"],
		StackName:            elementMap["StackName"],
		ClientRequestToken:   elementMap["ClientRequestToken"],
		Timestamp:            time,
	}, nil
}

func parseDate(str *string) (*time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, *str)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func main() {
	lambda.Start(handler)
}
