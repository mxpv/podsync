package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mxpv/podsync/pkg/model"
)

func TestCompress(t *testing.T) {
	inData := []model.Item{
		{ID: "1", Title: "title1"},
		{ID: "2", Title: "title2"},
	}

	data, err := compressObj(inData)
	assert.NoError(t, err)

	var outData []model.Item
	err = decompressObj(data, &outData)
	assert.NoError(t, err)

	assert.ObjectsAreEqual(inData, outData)
}
