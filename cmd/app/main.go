package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/proxy"
	"github.com/go-pg/pg"
	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/builders"
	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/feeds"
	"github.com/mxpv/podsync/pkg/handler"
	"github.com/mxpv/podsync/pkg/storage"
	"github.com/mxpv/podsync/pkg/support"
	"github.com/pkg/errors"
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

	redis, err := storage.NewRedisStorage(cfg.RedisURL)
	if err != nil {
		panic(err)
	}

	database, err := createPg(cfg.PostgresConnectionURL)
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
		feeds.WithStorage(redis),
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

	srv.Shutdown(ctx)
	database.Close()

	log.Printf("server gracefully stopped")
}

func createPg(connectionURL string) (*pg.DB, error) {
	opts, err := pg.ParseURL(connectionURL)
	if err != nil {
		return nil, err
	}

	// If host format is "projection:region:host", than use Google SQL Proxy
	// See https://github.com/go-pg/pg/issues/576
	if strings.Count(opts.Addr, ":") == 2 {
		log.Print("using GCP SQL proxy")
		opts.Dialer = func(network, addr string) (net.Conn, error) {
			return proxy.Dial(addr)
		}
	}

	db := pg.Connect(opts)

	// Check database connectivity
	if _, err := db.ExecOne("SELECT 1"); err != nil {
		db.Close()
		return nil, errors.Wrap(err, "failed to check database connectivity")
	}

	return db, nil
}
