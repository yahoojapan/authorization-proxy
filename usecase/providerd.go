/*
Copyright (C)  2018 Yahoo Japan Corporation Athenz team.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package usecase

import (
	"context"

	"github.com/kpango/glg"
	"github.com/pkg/errors"

	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/handler"
	"github.com/yahoojapan/authorization-proxy/infra"
	"github.com/yahoojapan/authorization-proxy/router"
	"github.com/yahoojapan/authorization-proxy/service"

	providerd "github.com/yahoojapan/athenz-authorizer"
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

	debugMux := router.NewDebugRouter(cfg.Server, handler.NewDebugHandler(athenz))

	return &providerDaemon{
		cfg:    cfg,
		athenz: athenz,
		server: service.NewServer(
			service.WithServerConfig(cfg.Server),
			service.WithServerHandler(handler.New(cfg.Proxy, infra.NewBuffer(cfg.Proxy.BufferSize), athenz)),
			service.WithDebugHandler(debugMux)),
	}, nil
}

// Start returns an error slice channel. This error channel reports the errors inside Authorization Proxy server.
func (g *providerDaemon) Start(ctx context.Context) <-chan []error {
	ech := make(chan []error)
	pch := g.athenz.Start(ctx)
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
		providerd.WithAthenzURL(cfg.Athenz.URL),
		providerd.WithPubkeyRefreshDuration(cfg.Authorization.PubKeyRefreshDuration),
		providerd.WithPubkeySysAuthDomain(cfg.Authorization.PubKeySysAuthDomain),
		providerd.WithPubkeyEtagExpTime(cfg.Authorization.PubKeyEtagExpTime),
		providerd.WithPubkeyEtagFlushDuration(cfg.Authorization.PubKeyEtagFlushDur),
		providerd.WithAthenzDomains(cfg.Authorization.AthenzDomains...),
		providerd.WithPolicyExpireMargin(cfg.Authorization.PolicyExpireMargin),
		providerd.WithPolicyRefreshDuration(cfg.Authorization.PolicyRefreshDuration),
		providerd.WithPolicyEtagFlushDuration(cfg.Authorization.PolicyEtagFlushDur),
		providerd.WithPolicyEtagExpTime(cfg.Authorization.PolicyEtagExpTime),
		providerd.WithDisableJwkd(),
	)
}
