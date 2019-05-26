package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	sess  = session.Must(session.NewSession())
	queue = sqs.New(sess)
)

const (
	chanSize           = 1024
	maxElementPerBatch = 10 // SQS Batch limit is 10 items per request
)

type Item struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Start   int    `json:"start"`
	Count   int    `json:"count"`
	LastID  string `json:"last_id"`
	Format  string `json:"format"`
	Quality string `json:"quality"`
}

type Sender struct {
	url    *string
	items  chan *Item
	cancel context.CancelFunc
}

func New(ctx context.Context, url string) *Sender {
	ctx, cancel := context.WithCancel(ctx)
	items := make(chan *Item, chanSize)

	sender := &Sender{
		url:    aws.String(url),
		items:  items,
		cancel: cancel,
	}

	go sender.transmit(ctx)

	return sender
}

func (s *Sender) Add(item *Item) {
	s.items <- item
}

func (s *Sender) Close() {
	s.cancel()
	close(s.items)
}

func (s *Sender) transmit(ctx context.Context) error {
	var list = make([]*Item, 0, maxElementPerBatch)

	flush := func(ctx context.Context) {
		if len(list) == 0 {
			return
		}

		if err := s.send(ctx, list); err != nil {
			log.WithError(err).Error("failed to send batch")
		}

		list = make([]*Item, 0, maxElementPerBatch)
	}

	for {
		select {
		case <-time.After(5 * time.Second):
			// Flush list if not filled up entirely within 5 seconds
			flush(ctx)

		case item := <-s.items:
			// Append an item to list and flush if filled up
			list = append(list, item)
			if len(list) == maxElementPerBatch {
				flush(ctx)
			}

		case <-ctx.Done():
			// Exiting, flush leftovers
			flush(context.Background())
			return ctx.Err()
		}
	}
}

func (s *Sender) send(ctx context.Context, list []*Item) error {
	if len(list) == 0 {
		return nil
	}

	log.Debugf("sending a new batch")

	sendInput := &sqs.SendMessageBatchInput{
		QueueUrl: s.url,
	}

	for _, item := range list {

		data, err := json.Marshal(item)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal item %q", item.ID)
		}

		body := string(data)

		sendInput.Entries = append(sendInput.Entries, &sqs.SendMessageBatchRequestEntry{
			Id:          aws.String(item.ID),
			MessageBody: aws.String(body),
		})

		log.Debugf("sending batch: %+v", sendInput)
	}

	_, err := queue.SendMessageBatchWithContext(ctx, sendInput)
	if err != nil {
		return errors.Wrap(err, "failed to send message batch")
	}

	log.Infof("sent %d item(s) to SQS", len(list))
	return nil
}
