package main

import (
	"context"
	"fmt"
	"log"
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
)

func main() {
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create core services

	cfg, err := config.ReadConfiguration()
	if err != nil {
		panic(err)
	}

	database, err := storage.NewPG(cfg.PostgresConnectionURL, true)
	if err != nil {
		panic(err)
	}

	statistics, err := stats.NewRedisStats(cfg.RedisURL)
	if err != nil {
		panic(err)
	}

	patreon := support.NewPatreon(database)

	// Builders

	youtube, err := builders.NewYouTubeBuilder(cfg.YouTubeApiKey)
	if err != nil {
		panic(err)
	}

	vimeo, err := builders.NewVimeoBuilder(ctx, cfg.VimeoApiKey)
	if err != nil {
		panic(err)
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
		log.Println("running listener")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	<-stop

	log.Printf("shutting down server")

	_ = srv.Shutdown(ctx)
	_ = database.Close()
	_ = statistics.Close()

	log.Printf("server gracefully stopped")
}
