package storage

import (
	"context"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	attr "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	expr "github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/pkg/errors"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"

	log "github.com/sirupsen/logrus"
)

const (
	pingTimeout       = 5 * time.Second
	pledgesPrimaryKey = "PatronID"
	feedsPrimaryKey   = "HashID"

	// Update LastAccess field every hour
	feedLastAccessUpdatePeriod = time.Hour
	feedTimeToLive             = time.Hour * 24 * 90
)

var (
	feedTimeToLiveField    = aws.String("ExpirationTime")
	feedDowngradeIndexName = aws.String("UserID-HashID-Index")
)

/*
Pledges:
	Table name:         Pledges
	Primary key:        PatronID (Number)
	RCU:                1 (used while creating a new feed)
	WCU:                1 (used when pledge changes)
	No secondary indexed needed
Feeds:
	Table name:         Feeds
	Primary key:        HashID (String)
	RCU:                10
	WCU:                5
	Secondary index:
		Primary key:    UserID (String)
		Sort key:       HashID (String)
		Index name:     UserID-HashID-Index
		Projected attr: Keys only
		RCU/WCU:        1/1
	TTL attr:           ExpirationTime
*/
type Dynamo struct {
	dynamo           *dynamodb.DynamoDB
	FeedsTableName   *string
	PledgesTableName *string
}

func NewDynamo(cfg ...*aws.Config) (Dynamo, error) {
	sess, err := session.NewSession(cfg...)
	if err != nil {
		return Dynamo{}, err
	}

	db := dynamodb.New(sess)

	// Verify connectivity
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	_, err = db.ListTablesWithContext(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return Dynamo{}, err
	}

	return Dynamo{
		dynamo:           db,
		FeedsTableName:   aws.String("Feeds"),
		PledgesTableName: aws.String("Pledges"),
	}, nil
}

func (d Dynamo) SaveFeed(feed *model.Feed) error {
	logger := log.WithFields(log.Fields{
		"hash_id": feed.HashID,
		"user_id": feed.UserID,
	})

	now := time.Now().UTC()

	feed.LastAccess = now
	feed.ExpirationTime = now.Add(feedTimeToLive)

	item, err := attr.MarshalMap(feed)
	if err != nil {
		logger.WithError(err).Error("failed to marshal feed model")
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName:           d.FeedsTableName,
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(HashID)"),
	}

	if _, err := d.dynamo.PutItem(input); err != nil {
		logger.WithError(err).Error("failed to save feed item")
		return err
	}

	return nil
}

func (d Dynamo) GetFeed(hashID string) (*model.Feed, error) {
	logger := log.WithField("hash_id", hashID)

	logger.Debug("getting feed")

	getInput := &dynamodb.GetItemInput{
		TableName: d.FeedsTableName,
		Key: map[string]*dynamodb.AttributeValue{
			"HashID": {S: aws.String(hashID)},
		},
	}

	getOutput, err := d.dynamo.GetItem(getInput)
	if err != nil {
		logger.WithError(err).Error("failed to get feed item")
		return nil, err
	}

	if getOutput.Item == nil {
		return nil, errors.New("not found")
	}

	var feed model.Feed
	if err := attr.UnmarshalMap(getOutput.Item, &feed); err != nil {
		logger.WithError(err).Error("failed to unmarshal feed item")
		return nil, err
	}

	// Check if we need to update LastAccess field (no more than once per hour)
	now := time.Now().UTC()
	if feed.LastAccess.Add(feedLastAccessUpdatePeriod).Before(now) {
		logger.Debugf("updating feed's last access timestamp")

		// Set LastAccess field to now
		// Set ExpirationTime field to now + feedTimeToLive
		updateExpression, err := expr.
			NewBuilder().
			WithUpdate(expr.
				Set(expr.Name("LastAccess"), expr.Value(now)).
				Set(expr.Name("ExpirationTime"), expr.Value(now.Add(feedTimeToLive)))).
			Build()

		if err != nil {
			logger.WithError(err).Error("failed to build update expression")
			return nil, err
		}

		updateInput := &dynamodb.UpdateItemInput{
			TableName:        d.FeedsTableName,
			Key:              getInput.Key,
			UpdateExpression: updateExpression.Update(),
		}

		_, err = d.dynamo.UpdateItem(updateInput)
		if err != nil {
			logger.WithError(err).Error("failed to update feed item")
			return nil, err
		}

		feed.LastAccess = now
	}

	return &feed, nil
}

func (d Dynamo) GetMetadata(hashID string) (*model.Feed, error) {
	logger := log.WithField("hash_id", hashID)

	logger.Debug("getting metadata")

	projectionExpression, err := expr.
		NewBuilder().
		WithProjection(
			expr.NamesList(
				expr.Name("FeedID"),
				expr.Name("HashID"),
				expr.Name("UserID"),
				expr.Name("Provider"),
				expr.Name("Format"),
				expr.Name("Quality"))).
		Build()

	if err != nil {
		logger.WithError(err).Error("failed to build projection expression")
		return nil, err
	}

	input := &dynamodb.GetItemInput{
		TableName: d.FeedsTableName,
		Key: map[string]*dynamodb.AttributeValue{
			"HashID": {S: aws.String(hashID)},
		},
		ProjectionExpression:     projectionExpression.Projection(),
		ExpressionAttributeNames: projectionExpression.Names(),
	}

	output, err := d.dynamo.GetItem(input)
	if err != nil {
		logger.WithError(err).Error("failed to get metadata item")
		return nil, err
	}

	if output.Item == nil {
		return nil, errors.New("not found")
	}

	var feed model.Feed
	if err := attr.UnmarshalMap(output.Item, &feed); err != nil {
		logger.WithError(err).Error("failed to unmarshal metadata item")
		return nil, err
	}

	return &feed, nil
}

func (d Dynamo) Downgrade(userID string, featureLevel int) error {
	logger := log.WithFields(log.Fields{
		"user_id":       userID,
		"feature_level": featureLevel,
	})

	logger.Debug("downgrading user's feeds")

	if featureLevel > api.ExtendedFeatures {
		// Max page size: 600
		// Format: any
		// Quality: any
		return nil
	}

	keyConditionExpression, err := expr.
		NewBuilder().
		WithKeyCondition(expr.KeyEqual(expr.Key("UserID"), expr.Value(userID))).
		Build()

	if err != nil {
		logger.WithError(err).Error("failed to build key condition")
		return err
	}

	// Query all feed's hash ids for specified

	logger.Debug("querying hash ids")

	queryInput := &dynamodb.QueryInput{
		TableName:                 d.FeedsTableName,
		IndexName:                 feedDowngradeIndexName,
		KeyConditionExpression:    keyConditionExpression.KeyCondition(),
		ExpressionAttributeNames:  keyConditionExpression.Names(),
		ExpressionAttributeValues: keyConditionExpression.Values(),
		Select:                    aws.String(dynamodb.SelectAllProjectedAttributes),
	}

	var keys []map[string]*dynamodb.AttributeValue
	err = d.dynamo.QueryPages(queryInput, func(output *dynamodb.QueryOutput, lastPage bool) bool {
		for _, item := range output.Items {
			keys = append(keys, map[string]*dynamodb.AttributeValue{
				feedsPrimaryKey: item[feedsPrimaryKey],
			})
		}

		return true
	})

	if err != nil {
		logger.WithError(err).Error("query failed")
		return err
	}

	logger.Debugf("got %d key(s)", len(keys))
	if len(keys) == 0 {
		return nil
	}

	if featureLevel == api.ExtendedFeatures {
		// Max page size: 150
		// Format: any
		// Quality: any
		updateExpression, err := expr.
			NewBuilder().
			WithUpdate(expr.
				Set(expr.Name("PageSize"), expr.Value(150)).
				Set(expr.Name("FeatureLevel"), expr.Value(api.ExtendedFeatures))).
			WithCondition(expr.
				Name("PageSize").GreaterThan(expr.Value(150))).
			Build()

		if err != nil {
			logger.WithError(err).Error("failed to build update expression")
			return err
		}

		for _, key := range keys {
			input := &dynamodb.UpdateItemInput{
				TableName:                 d.FeedsTableName,
				Key:                       key,
				ConditionExpression:       updateExpression.Condition(),
				UpdateExpression:          updateExpression.Update(),
				ExpressionAttributeNames:  updateExpression.Names(),
				ExpressionAttributeValues: updateExpression.Values(),
			}

			_, err := d.dynamo.UpdateItem(input)
			if err != nil {
				logger.WithError(err).Error("failed to update item")
				return err
			}
		}

	} else if featureLevel == api.DefaultFeatures {
		// Page size: 50
		// Format: video
		// Quality: high
		updateExpression, err := expr.
			NewBuilder().
			WithUpdate(expr.
				Set(expr.Name("PageSize"), expr.Value(50)).
				Set(expr.Name("FeatureLevel"), expr.Value(api.DefaultFeatures)).
				Set(expr.Name("Format"), expr.Value(api.FormatVideo)).
				Set(expr.Name("Quality"), expr.Value(api.QualityHigh))).
			Build()

		if err != nil {
			return err
		}

		for _, key := range keys {
			input := &dynamodb.UpdateItemInput{
				TableName:                 d.FeedsTableName,
				Key:                       key,
				UpdateExpression:          updateExpression.Update(),
				ExpressionAttributeNames:  updateExpression.Names(),
				ExpressionAttributeValues: updateExpression.Values(),
			}

			_, err := d.dynamo.UpdateItem(input)
			if err != nil {
				logger.WithError(err).Error("failed to update item")
				return err
			}
		}
	}

	logger.Info("successfully downgraded user's feeds")
	return nil
}

func (d Dynamo) AddPledge(pledge *model.Pledge) error {
	logger := log.WithFields(log.Fields{
		"pledge_id": pledge.PledgeID,
		"user_id":   pledge.PatronID,
	})

	item, err := attr.MarshalMap(pledge)
	if err != nil {
		logger.WithError(err).Error("failed to marshal pledge")
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName:           d.PledgesTableName,
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PatronID)"),
	}

	if _, err := d.dynamo.PutItem(input); err != nil {
		logger.WithError(err).Error("failed to put item")
		return err
	}

	return nil
}

func (d Dynamo) UpdatePledge(patronID string, pledge *model.Pledge) error {
	logger := log.WithFields(log.Fields{
		"pledge_id": pledge.PledgeID,
		"user_id":   patronID,
	})

	logger.Infof("updating pledge (new amount: %d)", pledge.AmountCents)

	builder := expr.
		Set(expr.Name("DeclinedSince"), expr.Value(pledge.DeclinedSince)).
		Set(expr.Name("AmountCents"), expr.Value(pledge.AmountCents)).
		Set(expr.Name("TotalHistoricalAmountCents"), expr.Value(pledge.TotalHistoricalAmountCents)).
		Set(expr.Name("OutstandingPaymentAmountCents"), expr.Value(pledge.OutstandingPaymentAmountCents)).
		Set(expr.Name("IsPaused"), expr.Value(pledge.IsPaused))

	updateExpression, err := expr.NewBuilder().WithUpdate(builder).Build()
	if err != nil {
		logger.WithError(err).Error("failed to build update expression")
		return err
	}

	input := &dynamodb.UpdateItemInput{
		TableName: d.PledgesTableName,
		Key: map[string]*dynamodb.AttributeValue{
			pledgesPrimaryKey: {N: aws.String(patronID)},
		},
		UpdateExpression:          updateExpression.Update(),
		ExpressionAttributeNames:  updateExpression.Names(),
		ExpressionAttributeValues: updateExpression.Values(),
	}

	if _, err := d.dynamo.UpdateItem(input); err != nil {
		logger.WithError(err).Error("failed to update pledge")
		return err
	}

	return nil
}

func (d Dynamo) DeletePledge(pledge *model.Pledge) error {
	logger := log.WithFields(log.Fields{
		"pledge_id": pledge.PledgeID,
		"user_id":   pledge.PatronID,
	})

	pk := strconv.FormatInt(pledge.PatronID, 10)
	logger.Infof("deleting pledge %s", pk)

	input := &dynamodb.DeleteItemInput{
		TableName: d.PledgesTableName,
		Key: map[string]*dynamodb.AttributeValue{
			pledgesPrimaryKey: {N: aws.String(pk)},
		},
	}

	if _, err := d.dynamo.DeleteItem(input); err != nil {
		logger.WithError(err).Error("failed to delete pledge")
		return err
	}

	return nil
}

func (d Dynamo) GetPledge(patronID string) (*model.Pledge, error) {
	logger := log.WithField("user_id", patronID)

	logger.Debug("getting pledge")

	input := &dynamodb.GetItemInput{
		TableName: d.PledgesTableName,
		Key: map[string]*dynamodb.AttributeValue{
			pledgesPrimaryKey: {N: aws.String(patronID)},
		},
	}

	output, err := d.dynamo.GetItem(input)
	if err != nil {
		logger.WithError(err).Error("failed to get pledge item")
		return nil, err
	}

	if output.Item == nil {
		return nil, errors.New("not found")
	}

	var pledge model.Pledge
	if err := attr.UnmarshalMap(output.Item, &pledge); err != nil {
		logger.WithError(err).Error("failed to unmarshal pledge item")
		return nil, err
	}

	return &pledge, nil
}

func (d Dynamo) Close() error {
	return nil
}
