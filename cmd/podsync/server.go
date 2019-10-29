package main

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/mxpv/podsync/pkg/config"
)

type Server struct {
	http.Server
}

func NewServer(cfg *config.Config) *Server {
	port := cfg.Server.Port
	if port == 0 {
		port = 8080
	}

	srv := Server{}

	srv.Addr = fmt.Sprintf(":%d", port)
	log.Debugf("using address: %s", srv.Addr)

	fs := http.FileServer(http.Dir(cfg.Server.DataDir))
	http.Handle("/", fs)

	return &srv
}
