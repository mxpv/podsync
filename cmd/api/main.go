package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/builders"
	"github.com/mxpv/podsync/pkg/cache"
	"github.com/mxpv/podsync/pkg/feeds"
	"github.com/mxpv/podsync/pkg/handler"
	"github.com/mxpv/podsync/pkg/storage"
	"github.com/mxpv/podsync/pkg/support"
)

type Opts struct {
	YouTubeAPIKey          string `long:"youtube-key" required:"true" env:"YOUTUBE_API_KEY"`
	VimeoAPIKey            string `long:"vimeo-key" required:"true" env:"VIMEO_API_KEY"`
	PatreonClientID        string `long:"patreon-client-id" required:"true" env:"PATREON_CLIENT_ID"`
	PatreonSecret          string `long:"patreon-secret" required:"true" env:"PATREON_SECRET"`
	PatreonRedirectURL     string `long:"patreon-redirect-url" required:"true" env:"PATREON_REDIRECT_URL"`
	PatreonWebhooksSecret  string `long:"patreon-webhook-secret" required:"true" env:"PATREON_WEBHOOKS_SECRET"`
	PostgresConnectionURL  string `long:"pg-url" env:"POSTGRES_CONNECTION_URL"`
	CookieSecret           string `long:"cookie-secret" required:"true" env:"COOKIE_SECRET"`
	DynamoFeedsTableName   string `long:"dynamo-feeds-table" env:"DYNAMO_FEEDS_TABLE_NAME"`
	DynamoPledgesTableName string `long:"dynamo-pledges-table" env:"DYNAMO_PLEDGES_TABLE_NAME"`
	RedisURL               string `long:"redis-url" required:"true" env:"REDIS_CONNECTION_URL"`
	UpdaterURL             string `long:"updater-url" required:"true" env:"UPDATER_URL"`
	Debug                  bool   `long:"debug" env:"DEBUG"`
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create core services

	var opts Opts
	if _, err := flags.Parse(&opts); err != nil {
		log.WithError(err).Fatal("failed to read configuration")
	}

	if opts.Debug {
		log.SetLevel(log.DebugLevel)
	}

	database, err := storage.NewDynamo()
	if err != nil {
		log.WithError(err).Fatal("failed to create database")
	}

	if opts.DynamoPledgesTableName != "" {
		database.PledgesTableName = aws.String(opts.DynamoPledgesTableName)
	}

	if opts.DynamoFeedsTableName != "" {
		database.FeedsTableName = aws.String(opts.DynamoFeedsTableName)
	}

	patreon := support.NewPatreon(database)

	// Cache

	redisCache, err := cache.NewRedisCache(opts.RedisURL)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize Redis cache")
	}

	// Builders

	youtube, err := builders.NewYouTubeBuilder(opts.YouTubeAPIKey)
	if err != nil {
		log.WithError(err).Fatal("failed to create YouTube builder")
	}

	vimeo, err := builders.NewVimeoBuilder(ctx, opts.VimeoAPIKey)
	if err != nil {
		log.WithError(err).Fatal("failed to create Vimeo builder")
	}

	generic := builders.NewRemote(opts.UpdaterURL)

	feed, err := feeds.NewFeedService(database, redisCache, map[api.Provider]feeds.Builder{
		api.ProviderYoutube: youtube,
		api.ProviderVimeo:   vimeo,
		api.ProviderGeneric: generic,
	})

	if err != nil {
		log.WithError(err).Fatal("failed to create feed service")
	}

	web := handler.New(feed, patreon, handler.Opts{
		CookieSecret:          opts.CookieSecret,
		PatreonClientID:       opts.PatreonClientID,
		PatreonSecret:         opts.PatreonSecret,
		PatreonRedirectURL:    opts.PatreonRedirectURL,
		PatreonWebhooksSecret: opts.PatreonWebhooksSecret,
	})

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", 5001),
		Handler: web,
	}

	go func() {
		log.Infof("running listener at %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.WithError(err).Error("failed to listen")
		}
	}()

	<-stop

	log.Info("shutting down server")

	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Error("server shutdown failed")
	}

	if err := redisCache.Close(); err != nil {
		log.WithError(err).Error("failed to close redis cache")
	}

	if err := database.Close(); err != nil {
		log.WithError(err).Error("failed to close database")
	}

	log.Info("server gracefully stopped")
}
