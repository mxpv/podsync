package feeds

import (
	"time"

	shortid "github.com/ventu-io/go-shortid"
)

type IDGen struct {
	sid *shortid.Shortid
}

func NewIDGen() (IDGen, error) {
	sid, err := shortid.New(1, shortid.DefaultABC, uint64(time.Now().UnixNano()))
	if err != nil {
		return IDGen{}, err
	}

	return IDGen{sid}, nil
}

func (id IDGen) Generate() (string, error) {
	return id.sid.Generate()
}
