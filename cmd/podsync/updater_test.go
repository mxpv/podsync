package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mxpv/podsync/pkg/config"
)

func TestUpdater_hostname(t *testing.T) {
	u := Updater{
		config: &config.Config{
			Server: config.Server{
				Hostname: "localhost",
				Port:     7979,
			},
		},
	}

	assert.Equal(t, "http://localhost", u.hostname())

	// Trim end slash
	u.config.Server.Hostname = "https://localhost:8080/"
	assert.Equal(t, "https://localhost:8080", u.hostname())
}
