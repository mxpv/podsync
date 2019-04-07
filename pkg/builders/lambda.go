package builders

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"

	"github.com/mxpv/podsync/pkg/model"
)

const (
	functionName  = "Updater"
	functionAlias = "PROD"
)

type responsePayload struct {
	LastID       string        `json:"last_id"`
	Episodes     []*model.Item `json:"episodes"`
	ErrorMessage string        `json:"errorMessage"`
}

// Lambda builder does incremental feed updates (see cmd/updater)
type Lambda struct {
	client *lambda.Lambda
}

func NewLambda(cfg ...*aws.Config) (*Lambda, error) {
	sess, err := session.NewSession(cfg...)
	if err != nil {
		return nil, err
	}

	client := lambda.New(sess)
	return &Lambda{client: client}, nil
}

func (l *Lambda) Build(feed *model.Feed) error {
	input := map[string]interface{}{
		"url":     feed.ItemURL,
		"start":   1,
		"count":   feed.PageSize,
		"last_id": feed.LastID,
		"format":  feed.Format,
		"quality": feed.Quality,
	}

	payload, err := json.Marshal(input)
	if err != nil {
		return errors.Wrap(err, "failed to serialize payload")
	}

	request := &lambda.InvokeInput{}
	request.SetPayload(payload)
	request.SetFunctionName(functionName)
	request.SetQualifier(functionAlias)

	response, err := l.client.Invoke(request)
	if err != nil {
		return err
	}

	var out responsePayload
	if err := json.Unmarshal(response.Payload, &out); err != nil {
		return errors.Wrap(err, "failed to deserialize lambda response")
	}

	if out.ErrorMessage != "" {
		return errors.Errorf("lambda error: %s", out.ErrorMessage)
	}

	feed.LastID = out.LastID
	feed.Episodes = append(out.Episodes, feed.Episodes...)
	feed.UpdatedAt = time.Now().UTC()

	return nil
}
