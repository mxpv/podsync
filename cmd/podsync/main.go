package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/ytdl"
)

type Opts struct {
	ConfigPath string `long:"config" short:"c" default:"config.toml" env:"PODSYNC_CONFIG_PATH"`
	Debug      bool   `long:"debug"`
	NoBanner   bool   `long:"no-banner"`
}

const banner = `
 _______  _______  ______   _______           _        _______ 
(  ____ )(  ___  )(  __  \ (  ____ \|\     /|( (    /|(  ____ \
| (    )|| (   ) || (  \  )| (    \/( \   / )|  \  ( || (    \/
| (____)|| |   | || |   ) || (_____  \ (_) / |   \ | || |      
|  _____)| |   | || |   | |(_____  )  \   /  | (\ \) || |      
| (      | |   | || |   ) |      ) |   ) (   | | \   || |      
| )      | (___) || (__/  )/\____) |   | |   | )  \  || (____/\
|/       (_______)(______/ \_______)   \_/   |/    )_)(_______/
`

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
	})

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

	if !opts.NoBanner {
		log.Info(banner)
	}

	log.WithFields(log.Fields{
		"version": version,
		"commit":  commit,
		"date":    date,
	}).Info("running podsync")

	downloader, err := ytdl.New(ctx)
	if err != nil {
		log.WithError(err).Fatal("youtube-dl error")
	}

	// Load TOML file
	log.Debugf("loading configuration %q", opts.ConfigPath)
	cfg, err := config.LoadConfig(opts.ConfigPath)
	if err != nil {
		log.WithError(err).Fatal("failed to load configuration file")
	}

	// Run updater thread
	log.Debug("creating updater")
	updater, err := NewUpdater(cfg, downloader)
	if err != nil {
		log.WithError(err).Fatal("failed to create updater")
	}

	c := cron.New(cron.WithChain(cron.SkipIfStillRunning(nil)))

	group.Go(func() error {
		// Shutdown web server
		defer func() {
			log.Info("shutting down cron")
			c.Stop()
		}()

		for _, feed := range cfg.Feeds {
			// backward capability
			// todo: remove that someday
			if feed.Schedule == "" {
				feed.Schedule = fmt.Sprintf("@every %s", feed.UpdatePeriod.String())
				log.WithFields(log.Fields{"feedId": feed.ID}).Warn("you must update config file and set field Schedule instead of UpdatePeriod")
			}
			_, err = c.AddFunc(feed.Schedule, func() {
				log.Debugf("adding %q to update queue", feed.URL)

				if false {
					if err := updater.Update(ctx, feed); err != nil {
						log.WithError(err).Errorf("failed to update feed: %s", feed.URL)
					}
				}
			})

			if err != nil {
				log.WithError(err).Fatalf("can't create cron task for feed: %s", feed.ID)
			}

			log.Debugf("-> %s (update every %s)", feed.URL, feed.UpdatePeriod)

			// Perform initial update after CLI restart
			_ = updater.Update(ctx, feed)
		}

		c.Start()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

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

	if err := group.Wait(); err != nil && (err != context.Canceled && err != http.ErrServerClosed) {
		log.WithError(err).Error("wait error")
	}

	log.Info("gracefully stopped")
}
