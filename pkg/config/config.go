package config

import (
	"strings"

	"github.com/spf13/viper"
)

const FileName = "podsync"

type AppConfig struct {
	YouTubeApiKey          string `yaml:"youtubeApiKey"`
	VimeoApiKey            string `yaml:"vimeoApiKey"`
	PatreonClientId        string `yaml:"patreonClientId"`
	PatreonSecret          string `yaml:"patreonSecret"`
	PatreonRedirectURL     string `yaml:"patreonRedirectUrl"`
	PatreonWebhooksSecret  string `json:"patreonWebhooksSecret"`
	PostgresConnectionURL  string `yaml:"postgresConnectionUrl"`
	RedisURL               string `yaml:"redisUrl"`
	CookieSecret           string `yaml:"cookieSecret"`
	AssetsPath             string `yaml:"assetsPath"`
	TemplatesPath          string `yaml:"templatesPath"`
	DynamoFeedsTableName   string `yaml:"dynamoFeedsTableName"`
	DynamoPledgesTableName string `yaml:"dynamoPledgesTableName"`
}

func ReadConfiguration() (cfg *AppConfig, err error) {
	viper.SetConfigName(FileName)

	// Configuration file
	viper.AddConfigPath(".")
	viper.AddConfigPath("/app/config/")

	// Env variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	envmap := map[string]string{
		"youtubeApiKey":          "YOUTUBE_API_KEY",
		"vimeoApiKey":            "VIMEO_API_KEY",
		"patreonClientId":        "PATREON_CLIENT_ID",
		"patreonSecret":          "PATREON_SECRET",
		"patreonRedirectUrl":     "PATREON_REDIRECT_URL",
		"patreonWebhooksSecret":  "PATREON_WEBHOOKS_SECRET",
		"postgresConnectionUrl":  "POSTGRES_CONNECTION_URL",
		"redisUrl":               "REDIS_CONNECTION_URL",
		"cookieSecret":           "COOKIE_SECRET",
		"assetsPath":             "ASSETS_PATH",
		"templatesPath":          "TEMPLATES_PATH",
		"dynamoFeedsTableName":   "DYNAMO_FEEDS_TABLE_NAME",
		"dynamoPledgesTableName": "DYNAMO_PLEDGES_TABLE_NAME",
	}

	for k, v := range envmap {
		viper.BindEnv(k, v)
	}

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return
		}

		// Ignore file not found error
		err = nil
	}

	cfg = &AppConfig{}

	viper.Unmarshal(cfg)
	return
}
