package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"

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

	// Create web server
	port := cfg.Server.Port
	if port == 0 {
		port = 8080
	}

	srv := http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	log.Debugf("using address %s", srv.Addr)

	// Run listener
	go func() {
		log.Infof("running listener at %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.WithError(err).Error("failed to listen")
		}
	}()

	<-stop

	log.Info("shutting down")

	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Error("server shutdown failed")
	}

	log.Info("gracefully stopped")
}
