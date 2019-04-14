package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimestamp_MarshalJSON(t *testing.T) {
	time1 := Timestamp(time.Now())

	data, err := time1.MarshalJSON()
	assert.NoError(t, err)

	time2 := Timestamp{}

	err = time2.UnmarshalJSON(data)
	assert.NoError(t, err)

	assert.EqualValues(t, time.Time(time1).Unix(), time.Time(time2).Unix())
}
