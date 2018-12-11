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
patreonWebhooksSecret: "10"
dynamoFeedsTableName: "11"
dynamoPledgesTableName: "12"
awsAccessKey: "13"
awsAccessSecret: "14"
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
	require.Equal(t, "10", cfg.PatreonWebhooksSecret)
	require.Equal(t, "11", cfg.DynamoFeedsTableName)
	require.Equal(t, "12", cfg.DynamoPledgesTableName)
	require.Equal(t, "13", cfg.AWSAccessKey)
	require.Equal(t, "14", cfg.AWSAccessSecret)
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
	os.Setenv("PATREON_WEBHOOKS_SECRET", "1010")
	os.Setenv("DYNAMO_FEEDS_TABLE_NAME", "1111")
	os.Setenv("DYNAMO_PLEDGES_TABLE_NAME", "1212")
	os.Setenv("AWS_ACCESS_KEY", "1313")
	os.Setenv("AWS_ACCESS_SECRET", "1414")

	cfg, err := ReadConfiguration()
	require.NoError(t, err)

	require.Equal(t, "11", cfg.YouTubeApiKey)
	require.Equal(t, "22", cfg.VimeoApiKey)
	require.Equal(t, "33", cfg.PatreonClientId)
	require.Equal(t, "44", cfg.PatreonSecret)
	require.Equal(t, "55", cfg.PostgresConnectionURL)
	require.Equal(t, "66", cfg.CookieSecret)
	require.Equal(t, "77", cfg.PatreonRedirectURL)
	require.Equal(t, "1010", cfg.PatreonWebhooksSecret)
	require.Equal(t, "1111", cfg.DynamoFeedsTableName)
	require.Equal(t, "1212", cfg.DynamoPledgesTableName)
	require.Equal(t, "1313", cfg.AWSAccessKey)
	require.Equal(t, "1414", cfg.AWSAccessSecret)
}
