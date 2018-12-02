package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/builders"
	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/feeds"
	"github.com/mxpv/podsync/pkg/handler"
	"github.com/mxpv/podsync/pkg/stats"
	"github.com/mxpv/podsync/pkg/storage"
	"github.com/mxpv/podsync/pkg/support"

	log "github.com/sirupsen/logrus"
)

func main() {
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create core services

	cfg, err := config.ReadConfiguration()
	if err != nil {
		log.WithError(err).Fatal("failed to read configuration")
	}

	database, err := storage.NewPG(cfg.PostgresConnectionURL, true)
	if err != nil {
		log.WithError(err).Fatal("failed to create pg")
	}

	statistics, err := stats.NewRedisStats(cfg.RedisURL)
	if err != nil {
		log.WithError(err).Fatal("failed to create redis")
	}

	patreon := support.NewPatreon(database)

	// Builders

	youtube, err := builders.NewYouTubeBuilder(cfg.YouTubeApiKey)
	if err != nil {
		log.WithError(err).Fatal("failed to create YouTube builder")
	}

	vimeo, err := builders.NewVimeoBuilder(ctx, cfg.VimeoApiKey)
	if err != nil {
		log.WithError(err).Fatal("failed to create Vimeo builder")
	}

	feed, err := feeds.NewFeedService(
		feeds.WithStorage(database),
		feeds.WithStats(statistics),
		feeds.WithBuilder(api.ProviderYoutube, youtube),
		feeds.WithBuilder(api.ProviderVimeo, vimeo),
	)

	if err != nil {
		panic(err)
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

	if err := database.Close(); err != nil {
		log.WithError(err).Error("failed to close database")
	}

	if err := statistics.Close(); err != nil {
		log.WithError(err).Error("failed to close stats storage")
	}

	log.Info("server gracefully stopped")
}
