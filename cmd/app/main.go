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
	"github.com/mxpv/podsync/pkg/id"
	"github.com/mxpv/podsync/pkg/server"
	"github.com/mxpv/podsync/pkg/storage"
)

func main() {
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create core sevices

	cfg, err := config.ReadConfiguration()
	if err != nil {
		panic(err)
	}

	hashIds, err := id.NewIdGenerator()
	if err != nil {
		panic(err)
	}

	redis, err := storage.NewRedisStorage(cfg.RedisURL)
	if err != nil {
		panic(err)
	}

	// Builders

	youtube, err := builders.NewYouTubeBuilder(cfg.YouTubeApiKey)
	if err != nil {
		panic(err)
	}

	vimeo, err := builders.NewVimeoBuilder(ctx, cfg.VimeoApiKey)
	if err != nil {
		panic(err)
	}

	feed := feeds.NewFeedService(
		feeds.WithIdGen(hashIds),
		feeds.WithStorage(redis),
		feeds.WithBuilder(api.Youtube, youtube),
		feeds.WithBuilder(api.Vimeo, vimeo),
	)

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: server.MakeHandlers(feed, cfg),
	}

	go func() {
		log.Println("running listener")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	<-stop

	log.Printf("shutting down server")

	srv.Shutdown(ctx)

	log.Printf("server gracefully stopped")
}
