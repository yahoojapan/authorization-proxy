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
	"crypto/tls"
	"net/http"
	"time"

	"github.com/kpango/glg"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/yahoojapan/authorization-proxy/v4/config"
	"github.com/yahoojapan/authorization-proxy/v4/handler"
	"github.com/yahoojapan/authorization-proxy/v4/infra"
	"github.com/yahoojapan/authorization-proxy/v4/router"
	"github.com/yahoojapan/authorization-proxy/v4/service"

	authorizerd "github.com/yahoojapan/athenz-authorizer/v5"
)

// AuthzProxyDaemon represents Authorization Proxy daemon behavior.
type AuthzProxyDaemon interface {
	Init(ctx context.Context) error
	Start(ctx context.Context) <-chan []error
}

type authzProxyDaemon struct {
	cfg        config.Config
	athenz     service.Authorizationd
	server     service.Server
	grpcServer service.Server
}

// New returns a Authorization Proxy daemon, or error occurred.
// The daemon contains a token service authentication and authorization server.
// This function will also initialize the mapping rules for the authentication and authorization check.
func New(cfg config.Config) (AuthzProxyDaemon, error) {
	athenz, err := newAuthzD(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot newAuthzD(cfg)")
	}

	var tlsCfg *tls.Config

	if cfg.Server.TLS.Enable {
		var err error
		tlsCfg, err = service.NewTLSConfig(cfg.Server.TLS)
		if err != nil {
			return nil, err
		}
	}

	debugMux := router.NewDebugRouter(cfg.Server, athenz)
	gh, closer := handler.NewGRPC(
		handler.WithProxyConfig(cfg.Proxy),
		handler.WithRoleTokenConfig(cfg.Authorization.RoleToken),
		handler.WithAuthorizationd(athenz),
		handler.WithTLSConfig(tlsCfg),
	)

	srv, err := service.NewServer(
		service.WithServerConfig(cfg.Server),
		service.WithRestHandler(handler.New(cfg.Proxy, infra.NewBuffer(cfg.Proxy.BufferSize), athenz)),
		service.WithDebugHandler(debugMux),
		service.WithGRPCHandler(gh),
		service.WithGRPCCloser(closer),
	)
	if err != nil {
		return nil, err
	}

	return &authzProxyDaemon{
		cfg:    cfg,
		athenz: athenz,
		server: srv,
	}, nil
}

// Init initializes child daemons synchronously.
func (g *authzProxyDaemon) Init(ctx context.Context) error {
	return g.athenz.Init(ctx)
}

// Start returns a channel of error slice . This error channel reports the errors inside the Authorizer daemon and the Authorization Proxy server.
func (g *authzProxyDaemon) Start(ctx context.Context) <-chan []error {
	ech := make(chan []error)
	var emap map[string]uint64 // used for returning value from child goroutine, should not r/w in this goroutine
	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)

	// handle authorizer daemon error, return on channel close
	eg.Go(func() error {
		// closure, only this goroutine should write on the variable and the map
		emap = make(map[string]uint64, 1)
		pch := g.athenz.Start(ctx)

		for err := range pch {
			if err != nil {
				glg.Errorf("pch: %v", err)
				// count errors by cause
				cause := errors.Cause(err).Error()
				_, ok := emap[cause]
				if !ok {
					emap[cause] = 1
				} else {
					emap[cause]++
				}
			}
		}

		return nil
	})

	// handle proxy server error, return on server shutdown done
	eg.Go(func() error {
		errs := <-g.server.ListenAndServe(ctx)
		glg.Errorf("sch: %v", errs)

		if len(errs) == 0 {
			// cannot be nil so that the context can cancel
			return errors.New("")
		}
		var baseErr error
		for i, err := range errs {
			if i == 0 {
				baseErr = err
			} else {
				baseErr = errors.Wrap(baseErr, err.Error())
			}
		}
		return baseErr
	})

	// wait for shutdown, and summarize errors
	go func() {
		defer close(ech)

		<-ctx.Done()
		err := eg.Wait()

		/*
			Read on emap is safe here, if and only if:
			1. emap is not used in the parenet goroutine
			2. the writer goroutine returns only if all errors are written, i.e. pch is closed
			3. this goroutine should wait for the writer goroutine to end, i.e. eg.Wait()
		*/
		// aggregate all errors as array
		perrs := make([]error, 0, len(emap))
		for errMsg, count := range emap {
			perrs = append(perrs, errors.WithMessagef(errors.New(errMsg), "authorizerd: %d times appeared", count))
		}

		// proxy server go func, should always return not nil error
		ech <- append(perrs, err)
	}()

	return ech
}

func newAuthzD(cfg config.Config) (service.Authorizationd, error) {
	client := http.DefaultClient
	if cfg.Athenz.Timeout != "" {
		t, err := time.ParseDuration(cfg.Athenz.Timeout)
		if err != nil {
			return nil, errors.Wrap(err, "newAuthzD(): Athenz.Timeout")
		}
		client.Timeout = t
	}
	if cfg.Athenz.CAPath != "" {
		cp, err := service.NewX509CertPool(cfg.Athenz.CAPath)
		if err != nil {
			return nil, errors.Wrap(err, "newAuthzD(): Athenz.CAPath")
		}
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: cp,
			},
		}
	}

	authzCfg := cfg.Authorization
	sharedOpts := []authorizerd.Option{
		authorizerd.WithAthenzURL(cfg.Athenz.URL),
		authorizerd.WithHTTPClient(client),
	}
	pubkeyOpts := []authorizerd.Option{
		authorizerd.WithPubkeySysAuthDomain(authzCfg.PublicKey.SysAuthDomain),
		authorizerd.WithPubkeyRefreshPeriod(authzCfg.PublicKey.RefreshPeriod),
		authorizerd.WithPubkeyETagExpiry(authzCfg.PublicKey.ETagExpiry),
		authorizerd.WithPubkeyETagPurgePeriod(authzCfg.PublicKey.ETagPurgePeriod),
		authorizerd.WithPubkeyRetryDelay(authzCfg.PublicKey.RetryDelay),
	}
	var policyOpts []authorizerd.Option
	if authzCfg.Policy.Disable {
		policyOpts = []authorizerd.Option{
			authorizerd.WithDisablePolicyd(),
		}
	} else {
		policyOpts = []authorizerd.Option{
			authorizerd.WithAthenzDomains(authzCfg.AthenzDomains...),
			authorizerd.WithPolicyExpiryMargin(authzCfg.Policy.ExpiryMargin),
			authorizerd.WithPolicyRefreshPeriod(authzCfg.Policy.RefreshPeriod),
			authorizerd.WithPolicyPurgePeriod(authzCfg.Policy.PurgePeriod),
			authorizerd.WithPolicyRetryDelay(authzCfg.Policy.RetryDelay),
			authorizerd.WithPolicyRetryAttempts(authzCfg.Policy.RetryAttempts),
		}

		if rules := authzCfg.Policy.MappingRules; rules != nil {
			translator, err := authorizerd.NewMappingRules(rules)
			if err != nil {
				return nil, errors.Wrap(err, "newAuthzD(): Failed to create a MappingRules")
			}
			policyOpts = append(policyOpts, authorizerd.WithTranslator(translator))
		}
	}
	var rtOpts []authorizerd.Option
	if authzCfg.RoleToken.Enable {
		rtOpts = []authorizerd.Option{
			authorizerd.WithEnableRoleToken(),
			authorizerd.WithRoleAuthHeader(cfg.Authorization.RoleToken.RoleAuthHeader),
		}
	} else {
		rtOpts = []authorizerd.Option{
			authorizerd.WithDisableRoleToken(),
		}
	}
	rcOpts := []authorizerd.Option{
		authorizerd.WithDisableRoleCert(),
	}

	var atOpts []authorizerd.Option
	var jwkOpts []authorizerd.Option
	if authzCfg.AccessToken.Enable {
		atOpts = []authorizerd.Option{
			authorizerd.WithAccessTokenParam(
				authorizerd.NewAccessTokenParam(
					authzCfg.AccessToken.Enable,
					authzCfg.AccessToken.VerifyCertThumbprint,
					authzCfg.AccessToken.CertBackdateDuration,
					authzCfg.AccessToken.CertOffsetDuration,
					authzCfg.AccessToken.VerifyClientID,
					authzCfg.AccessToken.AuthorizedClientIDs,
				),
			),
		}
		jwkOpts = []authorizerd.Option{
			authorizerd.WithEnableJwkd(),
			// use value in config.go in later version
			authorizerd.WithJwkRefreshPeriod(authzCfg.JWK.RefreshPeriod),
			authorizerd.WithJwkRetryDelay(authzCfg.JWK.RetryDelay),
			authorizerd.WithJwkURLs(authzCfg.JWK.URLs),
		}
	} else {
		atOpts = []authorizerd.Option{
			authorizerd.WithAccessTokenParam(
				authorizerd.NewAccessTokenParam(
					false,
					false,
					"0h",
					"0h",
					false,
					nil,
				),
			),
		}
		jwkOpts = []authorizerd.Option{
			authorizerd.WithDisableJwkd(),
		}
	}

	authzOptss := [][]authorizerd.Option{
		sharedOpts,
		pubkeyOpts,
		policyOpts,
		rtOpts,
		rcOpts,
		atOpts,
		jwkOpts,
	}
	var authzOptsLen int
	for _, opts := range authzOptss {
		authzOptsLen += len(opts)
	}
	authzOpts := make([]authorizerd.Option, 0, authzOptsLen)
	for _, opts := range authzOptss {
		authzOpts = append(authzOpts, opts...)
	}
	return authorizerd.New(authzOpts...)
}
