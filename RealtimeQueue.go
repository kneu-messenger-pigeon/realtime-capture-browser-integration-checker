package main

import (
	"context"
	"fmt"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	dekanatEvents "github.com/kneu-messenger-pigeon/dekanat-events"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

type RealtimeQueue struct {
	client      *sqs.Client
	sqsQueueUrl *string
	t           *testing.T
}

func CreateRealtimeQueue(t *testing.T) *RealtimeQueue {
	keyPairMapping := [2][2]string{
		{"AWS_ACCESS_KEY_ID", "CONSUMER_AWS_ACCESS_KEY_ID"},
		{"AWS_SECRET_ACCESS_KEY", "CONSUMER_AWS_SECRET_ACCESS_KEY"},
	}
	backupsValues := [len(keyPairMapping)]string{}
	for index, keyPair := range keyPairMapping {
		backupsValues[index] = os.Getenv(keyPair[0])
		_ = os.Setenv(keyPair[0], os.Getenv(keyPair[1]))
	}

	// load config with overridden env vars
	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background())
	for index, keyPair := range keyPairMapping {
		_ = os.Setenv(keyPair[0], backupsValues[index])
	}

	assert.NoError(t, err, "awsConfig.LoadDefaultConfig(context.Background()) failed")

	client := sqs.NewFromConfig(awsCfg)

	queue := &RealtimeQueue{
		client:      client,
		sqsQueueUrl: &config.sqsQueueUrl,
		t:           t,
	}

	// clear queue
	gMInput := &sqs.ReceiveMessageInput{
		QueueUrl:            queue.sqsQueueUrl,
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     10,
	}
	var msgResult *sqs.ReceiveMessageOutput

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	for msgResult == nil || len(msgResult.Messages) != 0 {
		msgResult, err = queue.client.ReceiveMessage(ctx, gMInput)

		if !assert.NoError(t, err, "Failed to get message from SQS: %v \n", err) {
			return nil
		}

		for _, message := range msgResult.Messages {
			queue.Delete(message.ReceiptHandle)
		}

		fmt.Printf("Clear queue - deleted %d messages\n", len(msgResult.Messages))
	}

	return queue
}

func (queue *RealtimeQueue) Fetch(waitTime time.Duration) (event interface{}) {
	gMInput := &sqs.ReceiveMessageInput{
		QueueUrl:            queue.sqsQueueUrl,
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     int32(waitTime.Seconds()),
	}
	var err error
	var msgResult *sqs.ReceiveMessageOutput
	var message *dekanatEvents.Message

	ctx, cancel := context.WithTimeout(context.Background(), waitTime+time.Second*2)
	msgResult, err = queue.client.ReceiveMessage(ctx, gMInput)
	cancel()
	if err != nil {
		queue.t.Errorf("Failed to get message from SQS: %v \n", err)
		return nil
	}

	if msgResult == nil || len(msgResult.Messages) == 0 {
		return nil
	}

	message, err = dekanatEvents.CreateMessage(msgResult.Messages[0].Body, msgResult.Messages[0].ReceiptHandle)
	if err == nil {
		event, err = message.ToEvent()
	}

	queue.Delete(message.ReceiptHandle)

	if err == nil && event != nil {
		return event
	}

	queue.t.Errorf("Failed to decode Event message: %v \n%+v\n", err, message)

	return nil
}

func (queue *RealtimeQueue) Delete(receiptHandle *string) {
	dMInput := &sqs.DeleteMessageInput{
		QueueUrl:      queue.sqsQueueUrl,
		ReceiptHandle: receiptHandle,
	}

	_, err := queue.client.DeleteMessage(context.Background(), dMInput)
	assert.NoError(queue.t, err, "Failed to remove message %s: %v \n", *receiptHandle, err)
}

func (queue *RealtimeQueue) AssertNoOtherEvents(t *testing.T) {
	event := queue.Fetch(time.Second * 2)
	assert.Nil(t, event, "Unexpected event found")
}
