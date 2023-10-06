package main

import (
	"context"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	dekanatEvents "github.com/kneu-messenger-pigeon/dekanat-events"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
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

	return &RealtimeQueue{
		client:      client,
		sqsQueueUrl: &config.sqsQueueUrl,
		t:           t,
	}
}

func (queue *RealtimeQueue) Fetch(context context.Context) (event interface{}) {
	gMInput := &sqs.ReceiveMessageInput{
		QueueUrl:            queue.sqsQueueUrl,
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     20,
	}
	var err error
	var msgResult *sqs.ReceiveMessageOutput
	var message *dekanatEvents.Message

	for context.Err() == nil {
		msgResult, err = queue.client.ReceiveMessage(context, gMInput)
		if err != nil && context.Err() == nil {
			queue.t.Errorf("Failed to get message from SQS: %v \n", err)
			break
		}

		if msgResult == nil || len(msgResult.Messages) == 0 {
			continue
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
	}

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
