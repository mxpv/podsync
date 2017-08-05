package id

import (
	"hash/fnv"

	"github.com/mxpv/podsync/web/pkg/database"
	hd "github.com/speps/go-hashids"
)

const (
	minLength = 4
	salt      = "mVJIX8cDWQJ71oMw6xw9yYV9TA1rojDcKrhUaOqEfaE"
	alphabet  = "abcdefghijklmnopqrstuvwxyz1234567890"
)

type hashId struct {
	hid *hd.HashID
}

func hashString(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}

func (h *hashId) Encode(feed *database.Feed) (string, error) {
	// Don't create duplicate urls for same playlist/settings
	// https://github.com/podsync/issues/issues/6
	numbers := []int{
		hashString(feed.UserId),
		hashString(feed.URL),
		feed.PageSize,
		hashString(string(feed.Quality)),
		hashString(string(feed.Format)),
	}

	return h.hid.Encode(numbers)
}

func NewIdGenerator() (*hashId, error) {
	data := hd.NewData()
	data.MinLength = minLength
	data.Salt = salt
	data.Alphabet = alphabet
	hid := hd.NewWithData(data)
	return &hashId{hid}, nil
}
