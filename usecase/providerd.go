package usecase

import (
	"context"

	"github.com/pkg/errors"

	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/handler"
	"github.com/yahoojapan/authorization-proxy/infra"
	"github.com/yahoojapan/authorization-proxy/service"

	providerd "github.com/yahoojapan/athenz-policy-updater"
)

// AuthorizationDaemon represents Authorization Proxy daemon behavior.
type AuthorizationDaemon interface {
	Start(ctx context.Context) chan []error
}

type providerDaemon struct {
	cfg    config.Config
	athenz service.Authorization
	server service.Server
}

// New returns a Authorization Proxy daemon, or error occurred.
// The daemon contains a token service authentication and authorization server.
// This function will also initialize the mapping rules for the authentication and authorization check.
func New(cfg config.Config) (AuthorizationDaemon, error) {
	athenz, err := newAuthorizationd(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot newAuthorizationd(cfg)")
	}

	return &providerDaemon{
		cfg:    cfg,
		athenz: athenz,
		server: service.NewServer(cfg.Server, handler.New(cfg.Proxy, infra.NewBuffer(cfg.Proxy.BufferSize), athenz)),
	}, nil
}

// Start returns an error slice channel. This error channel reports the errors inside Authorization Proxy server.
func (g *providerDaemon) Start(ctx context.Context) chan []error {
	g.athenz.StartAuthorizationd(ctx)
	return g.server.ListenAndServe(ctx)
}

func newAuthorizationd(cfg config.Config) (providerd.Authorizationd, error) {
	return providerd.New(
		providerd.AthenzURL(cfg.Athenz.URL),

		providerd.AthenzConfRefreshDuration(cfg.Authorization.AthenzConfRefreshDuration),
		providerd.AthenzConfSysAuthDomain(cfg.Authorization.AthenzConfSysAuthDomain),
		providerd.AthenzConfEtagExpTime(cfg.Authorization.AthenzConfEtagExpTime),
		providerd.AthenzConfEtagFlushDur(cfg.Authorization.AthenzConfEtagFlushDur),

		providerd.AthenzDomains(cfg.Authorization.AthenzDomains),
		providerd.PolicyRefreshDuration(cfg.Authorization.PolicyExpireMargin),
		providerd.PolicyRefreshDuration(cfg.Authorization.PolicyRefreshDuration),
		providerd.PolicyEtagFlushDur(cfg.Authorization.PolicyEtagFlushDur),
		providerd.PolicyEtagExpTime(cfg.Authorization.PolicyEtagExpTime),
	)
}
