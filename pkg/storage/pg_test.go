package storage

import (
	"testing"
	"time"

	"github.com/go-pg/pg"
	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
)

var (
	testPledge = &model.Pledge{PledgeID: 12345, AmountCents: 400, PatronID: 1, CreatedAt: time.Now()}
	testFeed   = &model.Feed{FeedID: 1, HashID: "3", UserID: "3", ItemID: "4", LinkType: api.LinkTypeChannel, Provider: api.ProviderVimeo, Format: api.FormatAudio ,Quality: api.QualityLow}
)

func TestPostgres_SaveFeed(t *testing.T) {
	stor := createPG(t)
	defer func() { _ = stor.Close() }()

	err := stor.SaveFeed(testFeed)
	require.NoError(t, err)

	find := &model.Feed{FeedID: 1}
	err = stor.db.Model(find).Select()
	require.NoError(t, err)

	require.Equal(t, testFeed.FeedID, find.FeedID)
	require.Equal(t, testFeed.HashID, find.HashID)
	require.Equal(t, testFeed.UserID, find.UserID)
	require.Equal(t, testFeed.ItemID, find.ItemID)
	require.Equal(t, testFeed.LinkType, find.LinkType)
	require.Equal(t, testFeed.Provider, find.Provider)
}

func TestPostgres_GetFeed(t *testing.T) {
	stor := createPG(t)
	defer func() { _ = stor.Close() }()

	err := stor.SaveFeed(testFeed)
	require.NoError(t, err)

	find, err := stor.GetFeed(testFeed.HashID)
	require.NoError(t, err)

	require.Equal(t, testFeed.FeedID, find.FeedID)
	require.Equal(t, testFeed.HashID, find.HashID)
	require.Equal(t, testFeed.UserID, find.UserID)
	require.Equal(t, testFeed.ItemID, find.ItemID)
	require.Equal(t, testFeed.LinkType, find.LinkType)
	require.Equal(t, testFeed.Provider, find.Provider)
}

func TestService_UpdateLastAccess(t *testing.T) {
	stor := createPG(t)
	defer func() { _ = stor.Close() }()

	err := stor.db.Insert(testFeed)
	require.NoError(t, err)

	feed1, err := stor.GetFeed(testFeed.HashID)
	require.NoError(t, err)

	feed2, err := stor.GetFeed(testFeed.HashID)
	require.NoError(t, err)

	require.True(t, feed2.LastAccess.After(feed1.LastAccess))
}

func TestPostgres_GetMetadata(t *testing.T) {
	stor := createPG(t)
	defer func() { _ = stor.Close() }()

	err := stor.SaveFeed(testFeed)
	require.NoError(t, err)

	find, err := stor.GetMetadata(testFeed.HashID)
	require.NoError(t, err)

	require.Equal(t, testFeed.UserID, find.UserID)
	require.Equal(t, testFeed.Provider, find.Provider)
	require.Equal(t, testFeed.Quality, find.Quality)
	require.Equal(t, testFeed.Format, find.Format)
}

func TestService_DowngradeToAnonymous(t *testing.T) {
	stor := createPG(t)
	defer func() { _ = stor.Close() }()

	feed := &model.Feed{
		HashID:       "123456",
		UserID:       "123456",
		ItemID:       "123456",
		Provider:     api.ProviderVimeo,
		LinkType:     api.LinkTypeGroup,
		PageSize:     150,
		Quality:      api.QualityLow,
		Format:       api.FormatAudio,
		FeatureLevel: api.ExtendedFeatures,
	}

	err := stor.db.Insert(feed)
	require.NoError(t, err)

	err = stor.Downgrade(feed.UserID, api.DefaultFeatures)
	require.NoError(t, err)

	downgraded := &model.Feed{FeedID: feed.FeedID}
	err = stor.db.Select(downgraded)
	require.NoError(t, err)

	require.Equal(t, 50, downgraded.PageSize)
	require.Equal(t, api.QualityHigh, downgraded.Quality)
	require.Equal(t, api.FormatVideo, downgraded.Format)
	require.Equal(t, api.DefaultFeatures, downgraded.FeatureLevel)
}

func TestService_DowngradeToExtendedFeatures(t *testing.T) {
	stor := createPG(t)
	defer func() { _ = stor.Close() }()

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

	err := stor.db.Insert(feed)
	require.NoError(t, err)

	err = stor.Downgrade(feed.UserID, api.ExtendedFeatures)
	require.NoError(t, err)

	downgraded := &model.Feed{FeedID: feed.FeedID}
	err = stor.db.Select(downgraded)
	require.NoError(t, err)

	require.Equal(t, 150, downgraded.PageSize)
	require.Equal(t, feed.Quality, downgraded.Quality)
	require.Equal(t, feed.Format, downgraded.Format)
	require.Equal(t, api.ExtendedFeatures, downgraded.FeatureLevel)
}

func TestPostgres_AddPledge(t *testing.T) {
	stor := createPG(t)
	defer func() { _ = stor.Close() }()

	err := stor.AddPledge(testPledge)
	require.NoError(t, err)

	pledge := &model.Pledge{PledgeID: 12345}
	err = stor.db.Select(pledge)
	require.NoError(t, err)

	require.Equal(t, int64(12345), pledge.PledgeID)
	require.Equal(t, 400, pledge.AmountCents)
}

func TestPostgres_UpdatePledge(t *testing.T) {
	stor := createPG(t)
	defer func() { _ = stor.Close() }()

	err := stor.AddPledge(testPledge)
	require.NoError(t, err)

	err = stor.UpdatePledge("1", &model.Pledge{AmountCents: 999})
	require.NoError(t, err)

	pledge := &model.Pledge{PledgeID: 12345}
	err = stor.db.Select(pledge)
	require.NoError(t, err)
	require.Equal(t, 999, pledge.AmountCents)
}

func TestPostgres_DeletePledge(t *testing.T) {
	stor := createPG(t)
	defer func() { _ = stor.Close() }()

	err := stor.AddPledge(testPledge)
	require.NoError(t, err)

	err = stor.DeletePledge(testPledge)
	require.NoError(t, err)

	err = stor.db.Select(&model.Pledge{PledgeID: 12345})
	require.Equal(t, pg.ErrNoRows, err)
}

func TestPostgres_GetPledge(t *testing.T) {
	stor := createPG(t)
	defer func() { _ = stor.Close() }()

	err := stor.AddPledge(testPledge)
	require.NoError(t, err)

	pledge, err := stor.GetPledge("1")
	require.NoError(t, err)
	require.Equal(t, 400, pledge.AmountCents)
	require.Equal(t, int64(12345), pledge.PledgeID)
}

// docker run -it --rm -p 5432:5432 -e POSTGRES_DB=podsync postgres
func createPG(t *testing.T) Postgres {
	const localConnectionString = "postgres://postgres:@localhost/podsync?sslmode=disable"

	postgres, err := NewPG(localConnectionString, false)
	require.NoError(t, err)

	_, err = postgres.db.Exec(pgsql)
	require.NoError(t, err)

	for _, obj := range []interface{}{&model.Pledge{}, &model.Feed{}} {
		_, err = postgres.db.Model(obj).Where("1=1").Delete()
		require.NoError(t, err)
	}

	return postgres
}
