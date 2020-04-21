package feed

import (
	"sync"

	"github.com/pkg/errors"
)

type KeyProvider interface {
	Get() string
}

func NewKeyProvider(keys []string) (KeyProvider, error) {
	switch len(keys) {
	case 0:
		return nil, errors.New("no keys")
	case 1:
		return NewFixedKey(keys[0])
	default:
		return NewRotatedKeys(keys)
	}
}

type FixedKeyProvider struct {
	key string
}

func NewFixedKey(key string) (KeyProvider, error) {
	if key == "" {
		return nil, errors.New("key can't be empty")
	}

	return &FixedKeyProvider{key: key}, nil
}

func (p FixedKeyProvider) Get() string {
	return p.key
}

type RotatedKeyProvider struct {
	keys  []string
	lock  sync.Mutex
	index int
}

func NewRotatedKeys(keys []string) (KeyProvider, error) {
	if len(keys) < 2 {
		return nil, errors.Errorf("at least 2 keys required (got %d)", len(keys))
	}

	return &RotatedKeyProvider{keys: keys, index: 0}, nil
}

func (p *RotatedKeyProvider) Get() string {
	p.lock.Lock()
	defer p.lock.Unlock()

	current := p.index % len(p.keys)
	p.index++

	return p.keys[current]
}
