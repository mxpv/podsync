package storage

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/model"
)

func TestPostgres_UpdateLastAccess(t *testing.T) {
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

func TestPostgres(t *testing.T) {
	runStorageTests(t, func(t *testing.T) storage {
		return createPG(t)
	})
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
