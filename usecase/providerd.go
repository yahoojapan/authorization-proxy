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

// ProviderDaemon represents Provider Sidecar daemon behavior.
type ProviderDaemon interface {
	Start(ctx context.Context) chan []error
}

type providerDaemon struct {
	cfg    config.Config
	athenz service.Provider
	server service.Server
}

// New returns a Provider Sidecar daemon, or error occurred.
// The daemon contains a token service authentication and authorization server.
// This function will also initialize the mapping rules for the authentication and authorization check.
func New(cfg config.Config) (ProviderDaemon, error) {
	athenz, err := newProviderd(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot newProviderd(cfg)")
	}

	return &providerDaemon{
		cfg:    cfg,
		athenz: athenz,
		server: service.NewServer(cfg.Server, handler.New(cfg.Proxy, infra.NewBuffer(cfg.Proxy.BufferSize), athenz)),
	}, nil
}

// Start returns an error slice channel. This error channel reports the errors inside Provider Sidecar server.
func (g *providerDaemon) Start(ctx context.Context) chan []error {
	g.athenz.StartProviderd(ctx)
	return g.server.ListenAndServe(ctx)
}

func newProviderd(cfg config.Config) (providerd.Providerd, error) {
	return providerd.New(
		providerd.AthenzURL(cfg.Athenz.URL),

		providerd.AthenzConfRefreshDuration(cfg.Provider.AthenzConfRefreshDuration),
		providerd.AthenzConfSysAuthDomain(cfg.Provider.AthenzConfSysAuthDomain),
		providerd.AthenzConfEtagExpTime(cfg.Provider.AthenzConfEtagExpTime),
		providerd.AthenzConfEtagFlushDur(cfg.Provider.AthenzConfEtagFlushDur),

		providerd.AthenzDomains(cfg.Provider.AthenzDomains),
		providerd.PolicyRefreshDuration(cfg.Provider.PolicyExpireMargin),
		providerd.PolicyRefreshDuration(cfg.Provider.PolicyRefreshDuration),
		providerd.PolicyEtagFlushDur(cfg.Provider.PolicyEtagFlushDur),
		providerd.PolicyEtagExpTime(cfg.Provider.PolicyEtagExpTime),
	)
}
