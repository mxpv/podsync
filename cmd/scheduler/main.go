package main

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	expr "github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/sqs"
	log "github.com/sirupsen/logrus"
)

const (
	hashIDField          = "HashID"
	chanSize             = 1024
	maxElementPerBatch   = 10 // SQS Batch limit is 10 items per request
	maxParallelTransmits = 12
)

var (
	sess   = session.Must(session.NewSession())
	dynamo = dynamodb.New(sess)
	queue  = sqs.New(sess)

	// Lambda function configuration
	tableName = os.Getenv("SCHEDULER_DYNAMODB_TABLE_NAME")
	queueName = os.Getenv("SCHEDULER_SQS_QUEUE_URL")
)

func init() {
	if tableName == "" {
		panic("dynamodb table name not specified")
	}

	if queueName == "" {
		panic("sqs queue name not specified")
	}
}

func handler(ctx context.Context) error {
	ids, err := query(ctx)
	if err != nil {
		log.WithError(err).Error("failed to get query channel")
		return err
	}

	var (
		total int64
		wg    sync.WaitGroup
		start = time.Now()
	)

	wg.Add(maxParallelTransmits)

	for i := 0; i < maxParallelTransmits; i++ {
		go func() {
			transmit(ctx, ids, &total)
			wg.Done()
		}()
	}

	wg.Wait()

	log.Infof("successfully submitted %d entries to SQS (scan took: %s)", total, time.Now().Sub(start))
	return nil
}

// query scans DynamoDB table and pushes ids to a channel
func query(ctx context.Context) (<-chan string, error) {
	var (
		projection = expr.NamesList(expr.Name(hashIDField))
		builder    = expr.NewBuilder().WithProjection(projection)
	)

	expression, err := builder.Build()
	if err != nil {
		log.WithError(err).Error("failed to build projection expression")
		return nil, err
	}

	scanInput := &dynamodb.ScanInput{
		TableName:                 aws.String(tableName),
		ProjectionExpression:      expression.Projection(),
		ExpressionAttributeNames:  expression.Names(),
		ExpressionAttributeValues: expression.Values(),
	}

	out := make(chan string, chanSize)

	go func() {
		defer close(out)

		err = dynamo.ScanPagesWithContext(ctx, scanInput, func(page *dynamodb.ScanOutput, lastPage bool) bool {
			for _, item := range page.Items {
				hashID := *item[hashIDField].S
				out <- hashID
			}

			return true
		})

		if err != nil {
			log.WithError(err).Error("scan failed")
		}
	}()

	return out, nil
}

// transmit reads ids from channel into buffer and sends to SQS in 10 item batches
func transmit(ctx context.Context, ids <-chan string, counter *int64) {
	var list = make([]string, 0, maxElementPerBatch)

	for id := range ids {
		list = append(list, id)

		// Send batch
		if len(list) == maxElementPerBatch {
			if err := send(ctx, list, counter); err != nil {
				log.WithError(err).Error("failed to send batch")
			}

			list = make([]string, 0, maxElementPerBatch)
		}
	}

	if len(list) > 0 {
		if err := send(ctx, list, counter); err != nil {
			log.WithError(err).Error("failed to send batch")
		}
	}
}

// send enqueues list of items to SQS queue
func send(ctx context.Context, list []string, counter *int64) error {
	sendInput := &sqs.SendMessageBatchInput{
		QueueUrl: aws.String(queueName),
	}

	for _, id := range list {
		sendInput.Entries = append(sendInput.Entries, &sqs.SendMessageBatchRequestEntry{
			Id:                     aws.String(id),
			MessageBody:            aws.String(id),
			MessageDeduplicationId: aws.String(id),
			MessageGroupId:         aws.String("feeds"), // use one group for all ids
		})
	}

	_, err := queue.SendMessageBatchWithContext(ctx, sendInput)
	if err != nil {
		log.WithError(err).Error("failed to send batch to SQS")
		return err
	}

	var (
		batchSize = int64(len(list))
		newCount  = atomic.AddInt64(counter, batchSize)
	)

	log.Infof("submitted %d item(s) (total: %d)", batchSize, newCount)
	return nil
}

func main() {
	lambda.Start(handler)
}
