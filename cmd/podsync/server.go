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
	bindAddress := cfg.Server.BindAddress
	if bindAddress == "*" {
		bindAddress = ""
	}
	srv := Server{}

	srv.Addr = fmt.Sprintf("%s:%d", bindAddress, port)
	log.Debugf("using address: %s:%s", bindAddress, srv.Addr)

	fs := http.FileServer(http.Dir(cfg.Server.DataDir))
	path := cfg.Server.Path
	http.Handle(fmt.Sprintf("/%s", path), fs)
	log.Debugf("handle path: /%s", path)

	return &srv
}
