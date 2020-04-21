package config

import (
	"time"

	"github.com/pkg/errors"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	res, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}

	*d = Duration{res}
	return nil
}

// StringSlice is a toml extension that lets you to specify either a string
// value (a slice with just one element) or a string slice.
type StringSlice []string

func (s *StringSlice) UnmarshalTOML(decode func(interface{}) error) error {
	var single string
	if err := decode(&single); err == nil {
		*s = []string{single}
		return nil
	}

	var slice []string
	if err := decode(&slice); err == nil {
		*s = slice
		return nil
	}

	return errors.New("failed to decode string (slice) field")
}
