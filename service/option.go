package service

import (
	"io"
	"net/http"

	"github.com/yahoojapan/authorization-proxy/v4/config"
	"google.golang.org/grpc"
)

// Option represents a functional option
type Option func(*server)

// WithServerConfig returns a ServerConfig functional option
func WithServerConfig(cfg config.Server) Option {
	return func(s *server) {
		s.cfg = cfg
	}
}

// WithRestHandler returns a Rest Handler functional option
func WithRestHandler(h http.Handler) Option {
	return func(s *server) {
		s.srvHandler = h
	}
}

// WithRestHandler returns a gRPC Handler functional option
func WithGRPCHandler(h grpc.StreamHandler) Option {
	return func(s *server) {
		s.grpcHandler = h
	}
}

// WithGRPCCloser returns a gRPC closer functional option
func WithGRPCCloser(c io.Closer) Option {
	return func(s *server) {
		s.grpcCloser = c
	}
}

// WithGRPCServer returns a gRPC Server functional option
func WithGRPCServer(srv *grpc.Server) Option {
	return func(s *server) {
		s.grpcSrv = srv
	}
}

// WithDebugHandler returns a DebugHandler functional option
func WithDebugHandler(h http.Handler) Option {
	return func(s *server) {
		s.dsHandler = h
	}
}
