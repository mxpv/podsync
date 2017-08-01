package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const YamlConfig = `
youtubeApiKey: "1"
vimeoApiKey: "2"
patreonClientId: "3"
patreonSecret: "4"
postgresConnectionUrl: "5"
`

func TestReadYaml(t *testing.T) {
	defer viper.Reset()

	err := ioutil.WriteFile("./podsync.yaml", []byte(YamlConfig), 0644)
	defer os.Remove("./podsync.yaml")
	require.NoError(t, err)

	cfg, err := ReadConfiguration()
	require.NoError(t, err)

	require.Equal(t, "1", cfg.YouTubeApiKey)
	require.Equal(t, "2", cfg.VimeoApiKey)
	require.Equal(t, "3", cfg.PatreonClientId)
	require.Equal(t, "4", cfg.PatreonSecret)
	require.Equal(t, "5", cfg.PostgresConnectionURL)
}

func TestReadEnv(t *testing.T) {
	defer viper.Reset()
	defer os.Clearenv()

	os.Setenv("YOUTUBE_API_KEY", "11")
	os.Setenv("VIMEO_API_KEY", "22")
	os.Setenv("PATREON_CLIENT_ID", "33")
	os.Setenv("PATREON_SECRET", "44")
	os.Setenv("POSTGRES_CONNECTION_URL", "55")

	cfg, err := ReadConfiguration()
	require.NoError(t, err)

	require.Equal(t, "11", cfg.YouTubeApiKey)
	require.Equal(t, "22", cfg.VimeoApiKey)
	require.Equal(t, "33", cfg.PatreonClientId)
	require.Equal(t, "44", cfg.PatreonSecret)
}
