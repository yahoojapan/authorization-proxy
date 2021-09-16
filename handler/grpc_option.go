package handler

import (
	"crypto/tls"

	"github.com/yahoojapan/authorization-proxy/v4/config"
	"github.com/yahoojapan/authorization-proxy/v4/service"
)

// Option represents a functional option for gRPC Handler
type GRPCOption func(*GRPCHandler)

var defaultGRPCOptions = []GRPCOption{}

// WithProxyConfig returns a proxy config functional option
func WithProxyConfig(cfg config.Proxy) GRPCOption {
	return func(h *GRPCHandler) {
		h.proxyCfg = cfg
	}
}

// WithRoleTokenConfig returns a role token config functional option
func WithRoleTokenConfig(cfg config.RoleToken) GRPCOption {
	return func(h *GRPCHandler) {
		h.roleCfg = cfg
	}
}

// WithAuthorizationd returns a authorizationd functional option
func WithAuthorizationd(a service.Authorizationd) GRPCOption {
	return func(h *GRPCHandler) {
		h.authorizationd = a
	}
}

// WithTLSConfig returns a TLS config functional option
func WithTLSConfig(cfg *tls.Config) GRPCOption {
	return func(g *GRPCHandler) {
		g.tlsCfg = cfg
	}
}
