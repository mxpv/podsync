//go:generate mockgen -source=patreon.go -destination=patreon_mock_test.go -package=support

package support

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	patreon "github.com/mxpv/patreon-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
)

func TestToModel(t *testing.T) {
	pledge := createPledge()

	modelPledge, err := ToModel(pledge)
	require.NoError(t, err)

	require.Equal(t, modelPledge.PledgeID, int64(12345))
	require.Equal(t, modelPledge.AmountCents, 400)
	require.Equal(t, modelPledge.PatronID, int64(67890))
	require.NotNil(t, modelPledge.CreatedAt)
}

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pledge := createPledge()
	expected, _ := ToModel(pledge)

	storage := NewMockstorage(ctrl)
	storage.EXPECT().AddPledge(gomock.Eq(expected)).Times(1).Return(nil)

	hook := Patreon{db: storage}

	err := hook.Hook(pledge, patreon.EventCreatePledge)
	require.NoError(t, err)
}

func TestUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pledge := createPledge()
	expected, _ := ToModel(pledge)

	storage := NewMockstorage(ctrl)
	storage.EXPECT().UpdatePledge("67890", gomock.Eq(expected))

	hook := Patreon{db: storage}
	err := hook.Hook(pledge, patreon.EventUpdatePledge)
	require.NoError(t, err)
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pledge := createPledge()
	expected, _ := ToModel(pledge)

	storage := NewMockstorage(ctrl)
	storage.EXPECT().DeletePledge(expected)

	hook := Patreon{db: storage}
	err := hook.Hook(pledge, patreon.EventDeletePledge)
	require.NoError(t, err)
}

func TestFindPledge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expected := &model.Pledge{}

	storage := NewMockstorage(ctrl)
	storage.EXPECT().GetPledge("123").Times(1).Return(expected, nil)

	hook := Patreon{db: storage}
	res, err := hook.FindPledge("123")
	require.NoError(t, err)
	require.Equal(t, expected, res)
}

func TestGetFeatureLevel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pledge := createPledge()
	storage := NewMockstorage(ctrl)

	ret, err := ToModel(pledge)
	require.NoError(t, err)

	storage.EXPECT().GetPledge(pledge.Relationships.Patron.Data.ID).Return(ret, nil)
	storage.EXPECT().GetPledge("xyz").Return(nil, errors.New("not found"))

	hook := Patreon{db: storage}

	require.Equal(t, api.PodcasterFeatures, hook.GetFeatureLevelByID(creatorID))
	require.Equal(t, api.DefaultFeatures, hook.GetFeatureLevelByID("xyz"))
	require.Equal(t, api.ExtendedPagination, hook.GetFeatureLevelByID(pledge.Relationships.Patron.Data.ID))
}

func createPledge() *patreon.Pledge {
	pledge := &patreon.Pledge{
		ID:   "12345",
		Type: "pledge",
	}

	pledge.Attributes.AmountCents = 400
	pledge.Attributes.CreatedAt = patreon.NullTime{Valid: true, Time: time.Now().UTC()}

	pledge.Relationships.Patron = &patreon.PatronRelationship{}
	pledge.Relationships.Patron.Data.ID = "67890"

	return pledge
}
