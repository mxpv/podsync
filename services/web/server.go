package web

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	http.Server
}

type Config struct {
	// Hostname to use for download links
	Hostname string `toml:"hostname"`
	// Port is a server port to listen to
	Port int `toml:"port"`
	// Bind a specific IP addresses for server
	// "*": bind all IP addresses which is default option
	// localhost or 127.0.0.1  bind a single IPv4 address
	BindAddress string `toml:"bind_address"`
	// Flag indicating if the server will use TLS
	TLS bool `toml:"tls"`
	// Path to a certificate file for TLS connections
	CertificatePath string `toml:"certificate_path"`
	// Path to a private key file for TLS connections
	KeyFilePath string `toml:"key_file_path"`
	// Specify path for reverse proxy and only [A-Za-z0-9]
	Path string `toml:"path"`
	// DataDir is a path to a directory to keep XML feeds and downloaded episodes,
	// that will be available to user via web server for download.
	DataDir string `toml:"data_dir"`
	// WebUIEnabled is a flag indicating if web UI is enabled
	WebUIEnabled bool `toml:"web_ui"`
}

func New(cfg Config, storage http.FileSystem) *Server {
	port := cfg.Port
	if port == 0 {
		port = 8080
	}

	bindAddress := cfg.BindAddress
	if bindAddress == "*" {
		bindAddress = ""
	}

	srv := Server{}

	srv.Addr = fmt.Sprintf("%s:%d", bindAddress, port)
	log.Debugf("using address: %s:%s", bindAddress, srv.Addr)

	fileServer := http.FileServer(storage)

	log.Debugf("handle path: /%s", cfg.Path)
	http.Handle(fmt.Sprintf("/%s", cfg.Path), fileServer)

	return &srv
}
