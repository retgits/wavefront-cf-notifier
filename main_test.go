package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	stackMessage = "StackId='arn:aws:cloudformation:us-west-2:123456789012:stack/MyStack/b9e8d9b0-be10-11e9-aa8d-0a1528792fcb'\nTimestamp='2019-08-13T21:24:52.887Z'\nEventId='c9fb6200-be10-11e9-9c1a-0621218a9930'\nLogicalResourceId='MyStack'\nNamespace='123456789012'\nPhysicalResourceId='arn:aws:cloudformation:us-west-2:123456789012:stack/MyStack/b9e8d9b0-be10-11e9-aa8d-0a1528792fcb'\nResourceProperties='null'\nResourceStatus='CREATE_COMPLETE'\nResourceStatusReason=''\nResourceType='AWS::CloudFormation::Stack'\nStackName='MyStack'\nClientRequestToken='Console-CreateStack-6b6e28ac-09ab-a7ee-9cf6-20865fb3953b'\n"
)

func TestParseDate(t *testing.T) {
	assert := assert.New(t)

	date := "2019-08-13T21:24:52.887Z"

	parsedDate, err := parseDate(&date)
	assert.NoError(err)
	assert.Equal(parsedDate.Year(), 2019)

	date = "helloworld"

	parsedDate, err = parseDate(&date)
	assert.Error(err)
}

func TestParseCFMessage(t *testing.T) {
	assert := assert.New(t)

	message, err := parseCFMessage(stackMessage)
	assert.NoError(err)
	assert.Equal(*message.StackName, "MyStack")

	message, err = parseCFMessage("hello\nworld")
	assert.Error(err)
}

func TestWavefrontEvent(t *testing.T) {
	assert := assert.New(t)

	evt := WavefrontEvent{}
	_, err := evt.Marshal()
	assert.NoError(err)

	stackMessage := "StackId='arn:aws:cloudformation:us-west-2:123456789012:stack/MyStack/b9e8d9b0-be10-11e9-aa8d-0a1528792fcb'\nTimestamp='2019-08-13T21:24:52.887Z'\nEventId='c9fb6200-be10-11e9-9c1a-0621218a9930'\nLogicalResourceId='MyStack'\nNamespace='123456789012'\nPhysicalResourceId='arn:aws:cloudformation:us-west-2:123456789012:stack/MyStack/b9e8d9b0-be10-11e9-aa8d-0a1528792fcb'\nResourceProperties='null'\nResourceStatus='CREATE_COMPLETE'\nResourceStatusReason=''\nResourceType='AWS::CloudFormation::Stack'\nStackName='MyStack'\nClientRequestToken='Console-CreateStack-6b6e28ac-09ab-a7ee-9cf6-20865fb3953b'\n"
	stackEvent, _ := parseCFMessage(stackMessage)

	wavefrontEventURL = "hello"
	err = createWavefrontEvent(stackEvent)
	assert.Error(err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"fake payload"}`)
	}))
	defer ts.Close()

	wavefrontEventURL = ts.URL

	err = createWavefrontEvent(stackEvent)
	assert.NoError(err)

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, `{"fake payload"}`)
	}))
	defer ts.Close()

	wavefrontEventURL = ts.URL

	err = createWavefrontEvent(stackEvent)
	assert.Error(err)
}
