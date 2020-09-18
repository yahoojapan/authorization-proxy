package service

import (
	"net/http"

	"github.com/yahoojapan/authorization-proxy/v4/config"
)

// Option represents a functional option
type Option func(*server)

// WithServerConfig returns a ServerConfig functional option
func WithServerConfig(cfg config.Server) Option {
	return func(s *server) {
		s.cfg = cfg
	}
}

// WithServerHandler returns a ServerHandler functional option
func WithServerHandler(h http.Handler) Option {
	return func(s *server) {
		s.srvHandler = h
	}
}

// WithDebugHandler returns a DebugHandler functional option
func WithDebugHandler(h http.Handler) Option {
	return func(s *server) {
		s.dsHandler = h
	}
}
