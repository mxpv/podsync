package id

import (
	"hash/fnv"
	"time"

	"github.com/mxpv/podsync/pkg/model"
	"github.com/ventu-io/go-shortid"
)

type hashId struct {
	sid *shortid.Shortid
}

func hashString(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}

func (h *hashId) Generate(feed *model.Feed) (string, error) {
	return h.sid.Generate()
}

func NewIdGenerator() (*hashId, error) {
	sid, err := shortid.New(1, shortid.DefaultABC, uint64(time.Now().UnixNano()))
	if err != nil {
		return nil, err
	}

	return &hashId{sid}, nil
}
