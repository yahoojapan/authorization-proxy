package usecase

import (
	"context"

	"github.com/kpango/glg"
	"github.com/pkg/errors"

	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/handler"
	"github.com/yahoojapan/authorization-proxy/infra"
	"github.com/yahoojapan/authorization-proxy/service"

	providerd "github.com/yahoojapan/athenz-policy-updater"
)

// AuthorizationDaemon represents Authorization Proxy daemon behavior.
type AuthorizationDaemon interface {
	Start(ctx context.Context) <-chan []error
}

type providerDaemon struct {
	cfg    config.Config
	athenz service.Authorizationd
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
func (g *providerDaemon) Start(ctx context.Context) <-chan []error {
	ech := make(chan []error)
	pch := g.athenz.StartProviderd(ctx)
	sch := g.server.ListenAndServe(ctx)
	go func() {
		emap := make(map[error]uint64, 1)
		defer close(ech)

		for {
			select {
			case <-ctx.Done():
				errs := make([]error, 0, len(emap)+1)
				for err, count := range emap {
					errs = append(errs, errors.WithMessagef(err, "%d times appeared", count))
				}
				errs = append(errs, ctx.Err())
				ech <- errs
				return
			case e := <-pch:
				glg.Errorf("pch %v", e)
				glg.Error(e)
				cause := errors.Cause(e)
				_, ok := emap[cause]
				if !ok {
					emap[cause] = 0
				}
				emap[cause]++
			case errs := <-sch:
				glg.Errorf("sch %v", errs)
				ech <- errs
				return
			}
		}
	}()

	return ech
}

func newAuthorizationd(cfg config.Config) (service.Authorizationd, error) {
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
