package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/kpango/glg"
	"github.com/pkg/errors"

	"github.com/yahoojapan/authorization-proxy/config"
	providerd "github.com/yahoojapan/athenz-policy-updater"
)

// New creates a handler for handling different HTTP requests based on the given services. It also contains a reverse proxy for handling proxy request.
func New(cfg config.Proxy, bp httputil.BufferPool, prov providerd.Providerd) http.Handler {
	return &httputil.ReverseProxy{
		BufferPool: bp,
		Director: func(r *http.Request) {
			u := *r.URL
			u.Scheme = func() string {
				if cfg.Scheme != "" {
					return cfg.Scheme
				}
				return "http"
			}()
			u.Host = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
			req, err := http.NewRequest(r.Method, u.String(), r.Body)
			if err != nil {
				glg.Fatal(errors.Wrap(err, "NewRequest returned error"))
			}
			req.Header = r.Header
			*r = *req
		},
		Transport: &transport{
			prov:         prov,
			RoundTripper: &http.Transport{},
			cfg:          cfg,
		},
		ErrorHandler: func(rw http.ResponseWriter, r *http.Request, err error) {
			status := http.StatusUnauthorized

			// request context canceled
			if errors.Cause(err) == context.Canceled {
				status = http.StatusRequestTimeout
			}

			/*
				switch err {
				case providerd.ErrDenyByPolicy:
					status = http.StatusUnauthorized
				case providerd.ErrNoMatch:
					status = http.StatusBadRequest
				//case providerd.DENY_ROLETOKEN_EXPIRED:
				//	status = http.StatusUnauthorized
				case providerd.DENY_ROLETOKEN_INVALID:
				//	status = http.StatusUnauthorized
				case providerd.ErrDomainMismatch:
					status = http.StatusUnauthorized
				case providerd.ErrDomainNotFound:
					status = http.StatusUnauthorized
				//case providerd.DENY_DOMAIN_EXPIRED:
				//	status = http.StatusUnauthorized
				//case providerd.DENY_DOMAIN_EMPTY:
				//	status = http.StatusUnauthorized
				//case providerd.DENY_INVALID_PARAMETERS:
				//	status = http.StatusUnauthorized
				//case providerd.DENY_CERT_MISMATCH_ISSUER:
				//	status = http.StatusUnauthorized
				//case providerd.DENY_CERT_MISSING_SUBJECT:
				//	status = http.StatusUnauthorized
				//case providerd.DENY_CERT_MISSING_DOMAIN:
				//	status = http.StatusUnauthorized
				//case providerd.DENY_CERT_MISSING_ROLE_NAME:
				//	status = http.StatusUnauthorized
				case providerd.ErrInvalidPolicyResource:
					status = http.StatusUnauthorized
				case providerd.ErrInvalidToken:
					status = http.StatusUnauthorized
				case context.Canceled:
					status = HttpStatusClientClosedRequest
				case io.ErrUnexpectedEOF:
					status = HttpStatusClientClosedRequest
				default:
					// TODO
				}
			*/
			rw.WriteHeader(status)
			/*
				rw.Header().Set("Content-Type", ProblemJSONContentType)

					json.NewEncoder(rw).Encode(RFC7807WithAthenz{
						Type:          "",
						Title:         err.Error(),
						Status:        status,
						Detail:        err.Error(),
						Instance:      r.RequestURI,
						RoleToken:     r.Header.Get(cfg.RoleHeader),
						InvalidParams: []InvalidParam{},
					})
			*/
		},
	}
}
