package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/mxpv/podsync/pkg/config"
)

type Opts struct {
	Config string `long:"config" short:"c" default:"config.toml"`
	Debug  bool   `long:"debug" short:"d"`
}

func main() {
	log.SetFormatter(&log.TextFormatter{})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	group, ctx := errgroup.WithContext(ctx)

	// Parse args
	opts := Opts{}
	_, err := flags.Parse(&opts)
	if err != nil {
		log.WithError(err).Fatal("failed to parse command line arguments")
	}

	if opts.Debug {
		log.SetLevel(log.DebugLevel)
	}

	// Load TOML file
	log.Debugf("loading configuration %q", opts.Config)
	cfg, err := config.LoadConfig(opts.Config)
	if err != nil {
		log.WithError(err).Fatal("failed to load configuration file")
	}

	// Queue of feeds to update
	updates := make(chan *config.Feed, 16)
	defer close(updates)

	// Run updater thread
	updater, err := NewUpdater(cfg)
	if err != nil {
		log.WithError(err).Fatal("failed to create updater")
	}

	group.Go(func() error {
		for {
			select {
			case feed := <-updates:
				if err := updater.Update(ctx, feed); err != nil {
					log.WithError(err).Errorf("failed to update feed: %s", feed.URL)
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	// Run wait goroutines for each feed configuration
	for _, feed := range cfg.Feeds {
		_feed := feed
		group.Go(func() error {
			timer := time.NewTicker(_feed.UpdatePeriod.Duration)
			defer timer.Stop()

			for {
				select {
				case <-timer.C:
					updates <- _feed
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})
	}

	// Run web server
	srv := NewServer(cfg)

	group.Go(func() error {
		log.Infof("running listener at %s", srv.Addr)
		return srv.ListenAndServe()
	})

	group.Go(func() error {
		// Shutdown web server
		defer func() {
			log.Info("shutting down web server")
			if err := srv.Shutdown(ctx); err != nil {
				log.WithError(err).Error("server shutdown failed")
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-stop:
				cancel()
				return nil
			}
		}
	})

	if err := group.Wait(); err != nil && err != context.Canceled {
		log.WithError(err).Error("wait error")
	}

	log.Info("gracefully stopped")
}
