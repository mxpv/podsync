package storage

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
)

type storage interface {
	SaveFeed(feed *model.Feed) error
	GetFeed(hashID string) (*model.Feed, error)
	GetMetadata(hashID string) (*model.Feed, error)
	Downgrade(userID string, featureLevel int) error

	// Patreon pledges
	AddPledge(pledge *model.Pledge) error
	UpdatePledge(patronID string, pledge *model.Pledge) error
	DeletePledge(pledge *model.Pledge) error
	GetPledge(patronID string) (*model.Pledge, error)

	Close() error
}

var (
	testPledge = &model.Pledge{
		PledgeID:                      12345,
		AmountCents:                   400,
		PatronID:                      1,
		CreatedAt:                     time.Now().UTC(),
		TotalHistoricalAmountCents:    100,
		OutstandingPaymentAmountCents: 100,
		IsPaused:                      true,
	}

	testFeed = &model.Feed{
		FeedID:       1,
		HashID:       "3",
		UserID:       "4",
		ItemID:       "5",
		LinkType:     api.LinkTypeChannel,
		Provider:     api.ProviderVimeo,
		Format:       api.FormatAudio,
		Quality:      api.QualityLow,
		PageSize:     150,
		FeatureLevel: api.ExtendedFeatures,
		CreatedAt:    time.Now().UTC(),
		LastAccess:   time.Now().UTC(),
	}
)

func runStorageTests(t *testing.T, createFn func(t *testing.T) storage) {
	if testing.Short() {
		t.Skip("Skipping storage test in short mode")
	}

	// Feeds
	t.Run("SaveFeed", makeTest(createFn, testSaveFeed))
	t.Run("LastAccess", makeTest(createFn, testLastAccess))
	t.Run("GetMetadata", makeTest(createFn, testGetMetadata))
	t.Run("Downgrade", func(t *testing.T) {
		t.Run("DefaultFeatures", makeTest(createFn, testDowngradeToDefaultFeatures))
		t.Run("ExtendedFeatures", makeTest(createFn, testDowngradeToExtendedFeatures))
	})

	// Pledge tests
	t.Run("AddPledge", makeTest(createFn, testAddPledge))
	t.Run("GetPledge", makeTest(createFn, testGetPledge))
	t.Run("DeletePledge", makeTest(createFn, testDeletePledge))
	t.Run("UpdatePledge", makeTest(createFn, testUpdatePledge))
}

func makeTest(createFn func(t *testing.T) storage, testFn func(t *testing.T, storage storage)) func(t *testing.T) {
	return func(t *testing.T) {
		storage := createFn(t)

		testFn(t, storage)

		err := storage.Close()
		require.Nil(t, err)
	}
}

func testSaveFeed(t *testing.T, storage storage) {
	err := storage.SaveFeed(testFeed)
	require.NoError(t, err)

	find, err := storage.GetFeed(testFeed.HashID)
	require.NoError(t, err)

	require.Equal(t, testFeed.HashID, find.HashID)
	require.Equal(t, testFeed.UserID, find.UserID)
	require.Equal(t, testFeed.ItemID, find.ItemID)
	require.Equal(t, testFeed.LinkType, find.LinkType)
	require.Equal(t, testFeed.Provider, find.Provider)
}

func testGetMetadata(t *testing.T, storage storage) {
	err := storage.SaveFeed(testFeed)
	require.NoError(t, err)

	find, err := storage.GetMetadata(testFeed.HashID)
	require.NoError(t, err)

	require.Equal(t, testFeed.UserID, find.UserID)
	require.Equal(t, testFeed.Provider, find.Provider)
	require.Equal(t, testFeed.Quality, find.Quality)
	require.Equal(t, testFeed.Format, find.Format)

	require.Equal(t, 0, find.PageSize)
	require.Equal(t, time.Time{}.Unix(), find.CreatedAt.Unix())
	require.Equal(t, time.Time{}.Unix(), find.LastAccess.Unix())
	require.Equal(t, 0, find.FeatureLevel)
}

func testDowngradeToDefaultFeatures(t *testing.T, storage storage) {
	feed := &model.Feed{
		HashID:       "123456",
		UserID:       "123456",
		ItemID:       "123456",
		Provider:     api.ProviderVimeo,
		LinkType:     api.LinkTypeGroup,
		PageSize:     200,
		Quality:      api.QualityLow,
		Format:       api.FormatAudio,
		FeatureLevel: api.ExtendedFeatures,
	}

	err := storage.SaveFeed(feed)
	require.NoError(t, err)

	err = storage.Downgrade(feed.UserID, api.DefaultFeatures)
	require.NoError(t, err)

	downgraded, err := storage.GetFeed(feed.HashID)
	require.NoError(t, err)

	require.Equal(t, 50, downgraded.PageSize)
	require.Equal(t, api.QualityHigh, downgraded.Quality)
	require.Equal(t, api.FormatVideo, downgraded.Format)
	require.Equal(t, api.DefaultFeatures, downgraded.FeatureLevel)
}

func testDowngradeToExtendedFeatures(t *testing.T, storage storage) {
	feed := &model.Feed{
		HashID:       "123456",
		UserID:       "123456",
		ItemID:       "123456",
		Provider:     api.ProviderVimeo,
		LinkType:     api.LinkTypeGroup,
		PageSize:     500,
		Quality:      api.QualityLow,
		Format:       api.FormatAudio,
		FeatureLevel: api.ExtendedFeatures,
	}

	err := storage.SaveFeed(feed)
	require.NoError(t, err)

	err = storage.Downgrade(feed.UserID, api.ExtendedFeatures)
	require.NoError(t, err)

	downgraded, err := storage.GetFeed(feed.HashID)
	require.NoError(t, err)

	require.Equal(t, 150, downgraded.PageSize)
	require.Equal(t, feed.Quality, downgraded.Quality)
	require.Equal(t, feed.Format, downgraded.Format)
	require.Equal(t, api.ExtendedFeatures, downgraded.FeatureLevel)
}

func testLastAccess(t *testing.T, storage storage) {
	date := time.Now().AddDate(-1, 0, 0).UTC()

	feed := &model.Feed{
		FeedID:       1,
		HashID:       "3",
		UserID:       "4",
		ItemID:       "5",
		LinkType:     api.LinkTypeChannel,
		Provider:     api.ProviderVimeo,
		Format:       api.FormatAudio,
		Quality:      api.QualityLow,
		PageSize:     150,
		FeatureLevel: api.ExtendedFeatures,
		CreatedAt:    date,
		LastAccess:   date,
	}

	err := storage.SaveFeed(feed)
	require.NoError(t, err)

	result, err := storage.GetFeed(feed.HashID)
	require.NoError(t, err)

	require.True(t, result.LastAccess.Sub(time.Now().UTC()) < 2*time.Second)
}

func testAddPledge(t *testing.T, storage storage) {
	err := storage.AddPledge(testPledge)
	require.NoError(t, err)

	pledge, err := storage.GetPledge(strconv.FormatInt(testPledge.PatronID, 10))
	require.NoError(t, err)

	require.Equal(t, testPledge.PledgeID, pledge.PledgeID)
	require.Equal(t, testPledge.PatronID, pledge.PatronID)
	require.Equal(t, testPledge.CreatedAt.Unix(), pledge.CreatedAt.Unix())
	require.Equal(t, testPledge.DeclinedSince.Unix(), pledge.DeclinedSince.Unix())
	require.Equal(t, testPledge.AmountCents, pledge.AmountCents)
	require.Equal(t, testPledge.TotalHistoricalAmountCents, pledge.TotalHistoricalAmountCents)
	require.Equal(t, testPledge.OutstandingPaymentAmountCents, pledge.OutstandingPaymentAmountCents)
	require.Equal(t, testPledge.IsPaused, pledge.IsPaused)
}

func testGetPledge(t *testing.T, storage storage) {
	err := storage.AddPledge(testPledge)
	require.NoError(t, err)

	pledge, err := storage.GetPledge(strconv.FormatInt(testPledge.PatronID, 10))
	require.NoError(t, err)

	require.Equal(t, testPledge.PledgeID, pledge.PledgeID)
	require.Equal(t, testPledge.PatronID, pledge.PatronID)
	require.Equal(t, testPledge.CreatedAt.Unix(), pledge.CreatedAt.Unix())
	require.Equal(t, testPledge.DeclinedSince.Unix(), pledge.DeclinedSince.Unix())
	require.Equal(t, testPledge.AmountCents, pledge.AmountCents)
	require.Equal(t, testPledge.TotalHistoricalAmountCents, pledge.TotalHistoricalAmountCents)
	require.Equal(t, testPledge.OutstandingPaymentAmountCents, pledge.OutstandingPaymentAmountCents)
	require.Equal(t, testPledge.IsPaused, pledge.IsPaused)
}

func testDeletePledge(t *testing.T, storage storage) {
	err := storage.AddPledge(testPledge)
	require.NoError(t, err)

	err = storage.DeletePledge(testPledge)
	require.NoError(t, err)

	pledge, err := storage.GetPledge(strconv.FormatInt(testPledge.PatronID, 10))
	require.Error(t, err)
	require.Nil(t, pledge)
}

func testUpdatePledge(t *testing.T, storage storage) {
	err := storage.AddPledge(testPledge)
	require.NoError(t, err)

	now := time.Now().UTC()

	err = storage.UpdatePledge(strconv.FormatInt(testPledge.PatronID, 10), &model.Pledge{
		DeclinedSince:                 now,
		AmountCents:                   400,
		TotalHistoricalAmountCents:    800,
		OutstandingPaymentAmountCents: 900,
		IsPaused:                      true,
	})

	require.NoError(t, err)

	pledge, err := storage.GetPledge("1")
	require.NoError(t, err)

	require.Equal(t, testPledge.PledgeID, pledge.PledgeID)
	require.Equal(t, testPledge.PatronID, pledge.PatronID)
	require.Equal(t, testPledge.CreatedAt.Unix(), pledge.CreatedAt.Unix())
	require.Equal(t, now.Unix(), pledge.DeclinedSince.Unix())
	require.Equal(t, 400, pledge.AmountCents)
	require.Equal(t, 800, pledge.TotalHistoricalAmountCents)
	require.Equal(t, 900, pledge.OutstandingPaymentAmountCents)
	require.Equal(t, true, pledge.IsPaused)
}
