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
	"github.com/mxpv/podsync/pkg/db"
	"github.com/mxpv/podsync/pkg/fs"
	"github.com/mxpv/podsync/pkg/ytdl"

	"gopkg.in/natefinch/lumberjack.v2"
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

	// Load TOML file
	log.Debugf("loading configuration %q", opts.ConfigPath)
	cfg, err := config.LoadConfig(opts.ConfigPath)
	if err != nil {
		log.WithError(err).Fatal("failed to load configuration file")
	}

	if cfg.Log.Filename != "" {
		log.SetOutput(&lumberjack.Logger{
			Filename:   cfg.Log.Filename,
			MaxSize:    cfg.Log.MaxSize,
			MaxBackups: cfg.Log.MaxBackups,
			MaxAge:     cfg.Log.MaxAge,
			Compress:   cfg.Log.Compress,
		})
	}

	log.WithFields(log.Fields{
		"version": version,
		"commit":  commit,
		"date":    date,
	}).Info("running podsync")

	downloader, err := ytdl.New(ctx, cfg.Downloader.SelfUpdate)
	if err != nil {
		log.WithError(err).Fatal("youtube-dl error")
	}

	database, err := db.NewBadger(&cfg.Database)
	if err != nil {
		log.WithError(err).Fatal("failed to open database")
	}

	storage, err := fs.NewLocal(cfg.Server.DataDir, cfg.Server.Hostname)
	if err != nil {
		log.WithError(err).Fatal("failed to open storage")
	}

	// Run updater thread
	log.Debug("creating updater")
	updater, err := NewUpdater(cfg, downloader, database, storage)
	if err != nil {
		log.WithError(err).Fatal("failed to create updater")
	}

	// Queue of feeds to update
	updates := make(chan *config.Feed, 16)
	defer close(updates)

	// Create Cron
	c := cron.New(cron.WithChain(cron.SkipIfStillRunning(nil)))
	m := make(map[string]cron.EntryID)

	// Run updates listener
	group.Go(func() error {
		for {
			select {
			case feed := <-updates:
				if err := updater.Update(ctx, feed); err != nil {
					log.WithError(err).Errorf("failed to update feed: %s", feed.URL)
				} else {
					log.Infof("next update of %s: %s", feed.ID, c.Entry(m[feed.ID]).Next)
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	// Run cron scheduler
	group.Go(func() error {
		var cronID cron.EntryID

		for _, feed := range cfg.Feeds {
			if feed.CronSchedule == "" {
				feed.CronSchedule = fmt.Sprintf("@every %s", feed.UpdatePeriod.String())
			}
			_feed := feed
			if cronID, err = c.AddFunc(_feed.CronSchedule, func() {
				log.Debugf("adding %q to update queue", _feed.ID)
				updates <- _feed
			}); err != nil {
				log.WithError(err).Fatalf("can't create cron task for feed: %s", _feed.ID)
			}

			m[_feed.ID] = cronID
			log.Debugf("-> %s (update '%s')", _feed.ID, _feed.CronSchedule)
			// Perform initial update after CLI restart
			updates <- _feed
		}

		c.Start()

		for {
			<-ctx.Done()

			log.Info("shutting down cron")
			c.Stop()

			return ctx.Err()
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

	if err := database.Close(); err != nil {
		log.WithError(err).Error("failed to close database")
	}

	log.Info("gracefully stopped")
}
