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
cookieSecret: "6"
patreonRedirectUrl: "7"
assetsPath: "8"
templatesPath: "9"
patreonWebhooksSecret: "10"
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
	require.Equal(t, "6", cfg.CookieSecret)
	require.Equal(t, "7", cfg.PatreonRedirectURL)
	require.Equal(t, "8", cfg.AssetsPath)
	require.Equal(t, "9", cfg.TemplatesPath)
	require.Equal(t, "10", cfg.PatreonWebhooksSecret)
}

func TestReadEnv(t *testing.T) {
	defer viper.Reset()
	defer os.Clearenv()

	os.Setenv("YOUTUBE_API_KEY", "11")
	os.Setenv("VIMEO_API_KEY", "22")
	os.Setenv("PATREON_CLIENT_ID", "33")
	os.Setenv("PATREON_SECRET", "44")
	os.Setenv("POSTGRES_CONNECTION_URL", "55")
	os.Setenv("COOKIE_SECRET", "66")
	os.Setenv("PATREON_REDIRECT_URL", "77")
	os.Setenv("ASSETS_PATH", "88")
	os.Setenv("TEMPLATES_PATH", "99")
	os.Setenv("PATREON_WEBHOOKS_SECRET", "1010")

	cfg, err := ReadConfiguration()
	require.NoError(t, err)

	require.Equal(t, "11", cfg.YouTubeApiKey)
	require.Equal(t, "22", cfg.VimeoApiKey)
	require.Equal(t, "33", cfg.PatreonClientId)
	require.Equal(t, "44", cfg.PatreonSecret)
	require.Equal(t, "55", cfg.PostgresConnectionURL)
	require.Equal(t, "66", cfg.CookieSecret)
	require.Equal(t, "77", cfg.PatreonRedirectURL)
	require.Equal(t, "88", cfg.AssetsPath)
	require.Equal(t, "99", cfg.TemplatesPath)
	require.Equal(t, "1010", cfg.PatreonWebhooksSecret)
}
