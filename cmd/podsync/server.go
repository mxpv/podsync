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

func NewServer(cfg *config.Config, storage http.FileSystem) *Server {
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

	fileServer := http.FileServer(storage)

	log.Debugf("handle path: /%s", cfg.Server.Path)
	http.Handle(fmt.Sprintf("/%s", cfg.Server.Path), fileServer)

	return &srv
}
