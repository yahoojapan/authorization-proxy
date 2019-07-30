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

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

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

	debugMux := router.NewDebugRouter(cfg.Server, athenz)

	return &providerDaemon{
		cfg:    cfg,
		athenz: athenz,
		server: service.NewServer(
			service.WithServerConfig(cfg.Server),
			service.WithServerHandler(handler.New(cfg.Proxy, infra.NewBuffer(cfg.Proxy.BufferSize), athenz)),
			service.WithDebugHandler(debugMux)),
	}, nil
}

// Start returns a channel of error slice . This error channel reports the errors inside the Authorizer daemon and the Authorization Proxy server.
func (g *providerDaemon) Start(ctx context.Context) <-chan []error {
	ech := make(chan []error)
	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)
	eg.Go(func() error {
		pch := g.athenz.Start(ctx)

		var errs error
		var ebuf []error
		for {

			// TODO use emap to aggregate errors
			emap := make(map[string]uint64, 1)
			e, ok := <-pch
			if !ok { // handle channel close
				pch = nil
				ech <- ebuf
				return nil
			}
			if e != nil {
				ebuf = append(ebuf, e)
				errs = errors.Wrap(errs, e.Error())
			}
		}
		return nil
	})

	eg.Go(func() error {
		sch := <-g.server.ListenAndServe(ctx)
		if len(sch) != 0 {
			ech <- sch
		}
		var errs error
		for _, err := range sch {
			errs = errors.Wrap(errs, err.Error())
		}
		return errs
	})

	go func() {
		defer close(ech)

		<-ctx.Done()
		err := eg.Wait()
		if err != nil {
			ech <- []error{err}
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
		providerd.WithPubkeyErrRetryInterval(cfg.Authorization.PubKeyErrRetryInterval),
		providerd.WithAthenzDomains(cfg.Authorization.AthenzDomains...),

		providerd.WithPolicyExpireMargin(cfg.Authorization.PolicyExpireMargin),
		providerd.WithPolicyRefreshDuration(cfg.Authorization.PolicyRefreshDuration),
		providerd.WithPolicyEtagFlushDuration(cfg.Authorization.PolicyEtagFlushDur),
		providerd.WithPolicyEtagExpTime(cfg.Authorization.PolicyEtagExpTime),
		providerd.WithPolicyErrRetryInterval(cfg.Authorization.PolicyErrRetryInterval),

		providerd.WithDisableJwkd(),
	)
}
