package id

import (
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

func (h *hashId) Encode(x ...int) (string, error) {
	var d []int
	return h.hid.Encode(append(d, x...))
}

func NewIdGenerator() (*hashId, error) {
	data := hd.NewData()
	data.MinLength = minLength
	data.Salt = salt
	data.Alphabet = alphabet
	hid, err := hd.NewWithData(data)
	if err != nil {
		return nil, err
	}

	return &hashId{hid}, nil
}
