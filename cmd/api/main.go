package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"

	log "github.com/sirupsen/logrus"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/builders"
	"github.com/mxpv/podsync/pkg/cache"
	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/feeds"
	"github.com/mxpv/podsync/pkg/handler"
	"github.com/mxpv/podsync/pkg/storage"
	"github.com/mxpv/podsync/pkg/support"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create core services

	cfg, err := config.ReadConfiguration()
	if err != nil {
		log.WithError(err).Fatal("failed to read configuration")
	}

	database, err := storage.NewDynamo(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(cfg.AWSAccessKey, cfg.AWSAccessSecret, ""),
	})

	if err != nil {
		log.WithError(err).Fatal("failed to create database")
	}

	if cfg.DynamoPledgesTableName != "" {
		database.PledgesTableName = aws.String(cfg.DynamoPledgesTableName)
	}

	if cfg.DynamoFeedsTableName != "" {
		database.FeedsTableName = aws.String(cfg.DynamoFeedsTableName)
	}

	patreon := support.NewPatreon(database)

	// Cache

	redisCache, err := cache.NewRedisCache(cfg.RedisURL)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize Redis cache")
	}

	// Builders

	youtube, err := builders.NewYouTubeBuilder(cfg.YouTubeAPIKey)
	if err != nil {
		log.WithError(err).Fatal("failed to create YouTube builder")
	}

	vimeo, err := builders.NewVimeoBuilder(ctx, cfg.VimeoAPIKey)
	if err != nil {
		log.WithError(err).Fatal("failed to create Vimeo builder")
	}

	generic, err := builders.NewLambda()
	if err != nil {
		log.WithError(err).Fatal("failed to create Lambda builder")
	}

	feed, err := feeds.NewFeedService(database, redisCache, map[api.Provider]feeds.Builder{
		api.ProviderYoutube: youtube,
		api.ProviderVimeo:   vimeo,
		api.ProviderGeneric: generic,
	})

	if err != nil {
		log.WithError(err).Fatal("failed to create feed service")
	}

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", 5001),
		Handler: handler.New(feed, patreon, cfg),
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
